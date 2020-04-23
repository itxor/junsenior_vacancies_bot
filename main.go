package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	redis_helper "parser/helpers/redis"
	"strconv"
)

// Vacancy ...
type Vacancy struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Area struct {
		Place string `json:"name"`
	} `json:"area"`
	Salary struct {
		From     int    `json:"from"`
		To       int    `json:"to"`
		Currency string `json:"currency"`
		Gross    bool   `json:"gross"`
	} `json:"salary"`
	PublishedAt string `json:"published_at"`
	CreatedAt   string `json:"created_at"`
	Archived    bool   `json:"archived"`
	URL         string `json:"alternate_url"`
	Snippet     struct {
		Description  string `json:"responsibility"`
		Requirements string `json:"requirement"`
	} `json:"snippet"`
	Employer struct {
		Name string `json:"name"`
	} `json:"employer"`
}

// Items ...
type Items struct {
	Vacancies   []Vacancy `json:"items"`
	Pages       int       `json:"pages"`
	CurrentPage int       `json:"page"`
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func main() {
	redisClient, err := redis_helper.NewRedisClient()
	if err != nil {
		panic(err)
	}

	databaseUrl, exists := os.LookupEnv("DATABASE_URL")
	if !exists {
		log.Fatalln("database url is not set")
		panic("database url is not set")
	}
	databaseDriver, exists := os.LookupEnv("DATABASE_DRIVER")
	if !exists {
		log.Fatalln("database driver is not set")
		panic("database url is not set")
	}

	db, err := sql.Open(databaseDriver, databaseUrl)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback() // Игнорируется, если в последующем изменения зафиксированы через Commit

	data := "IF NOT EXISTS (SELECT * FROM vacancies WHERE id = (?))" +
		"BEGIN " +
		"INSERT INTO vacancies(" +
		"id, " +
		"name, " +
		"place, " +
		"salary_from, " +
		"salary_to, " +
		"salary_currency, " +
		"salary_gross, " +
		"publiched_at, " +
		"archived, " +
		"url, " +
		"employer_name" +
		") VALUES (?,?,?,?,?,?,?,?,?,?,?) " +
		"END"
	fmt.Println(data)

	stmt, err := tx.Prepare("IF NOT EXISTS (SELECT * FROM vacancies WHERE id = (?)) " +
		"INSERT INTO vacancies(" +
		"id, " +
		"name, " +
		"place, " +
		"salary_from, " +
		"salary_to, " +
		"salary_currency, " +
		"salary_gross, " +
		"publiched_at, " +
		"archived, " +
		"url, " +
		"employer_name" +
		") VALUES (?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	lastUpdateTime := redisClient.GetRedisTimeStamp()
	vacancies, err := getVacansies(lastUpdateTime)
	if err != nil {
		panic(err)
	}
	redisClient.SetRedisTimeStamp()

	for _, vacancy := range vacancies.Vacancies {
		if _, err := stmt.Exec(
			vacancy.ID, // для поиска уже существующего значения
			vacancy.ID, // для записи в базу
			vacancy.Name,
			vacancy.Area.Place,
			vacancy.Salary.From,
			vacancy.Salary.To,
			vacancy.Salary.Currency,
			vacancy.Salary.Gross,
			vacancy.PublishedAt,
			vacancy.Archived,
			vacancy.URL,
			vacancy.Employer.Name,
		); err != nil {
			log.Fatal(err)
		}
	}

	if err := tx.Commit(); err != nil {
		fmt.Printf("Ошибка: %s", err.Error())
		panic(err)
	}
}

// получает вакансии и мапит их в Items-структуру
func getVacansies(lastUpdateTime string) (*Items, error) {
	resp, err := sendRequest(0, lastUpdateTime)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)

		return nil, err
	}

	var vacancies Items
	_ = json.Unmarshal(body, &vacancies)

	if vacancies.CurrentPage != 0 {
		return nil, errors.New("Невозможно распрасить страницу!")
	}
	countPages := vacancies.Pages

	var tempItems Items
	for i := 1; i < countPages; i++ {
		resp, err := sendRequest(i, lastUpdateTime)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)

			return nil, err
		}

		_ = json.Unmarshal(body, &tempItems)

		vacancies.Vacancies = append(vacancies.Vacancies, tempItems.Vacancies...)
	}

	return &vacancies, nil
}

// Отправляет запрос к hh на получение вакансий
func sendRequest(page int, dateFrom string) (*http.Response, error) {
	url, exists := os.LookupEnv("HH_URL")
	if !exists {
		log.Fatalln("head hanter base url not found")

		return nil, errors.New("head hanter base url not found")
	}

	req, _ := http.NewRequest(
		"GET",
		url,
		nil,
	)
	req.Header.Add("User-Agent", "api-test-agent")

	q := req.URL.Query()
	// запрос на слова в названии или описании вакансии
	q.Add("text", "NAME:(PHP OR Symfony OR Laravel) OR DESCRIPTION:(PHP OR Symfony OR Laravel)")
	q.Add("employment", "full")       // тип занятости - полная
	q.Add("employment", "part")       // или частичная
	q.Add("schedule", "remote")       // тип работы - удалённая
	q.Add("page", strconv.Itoa(page)) // страница
	if dateFrom != "" {
		q.Add("date_from", dateFrom) // ограничивает дату снизу
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)

		return nil, err
	}

	return resp, nil
}
