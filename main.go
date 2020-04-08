package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// Vacancy ...
type Vacancy struct {
	Id   int    `json:"id"`
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
	Url         string `json:"alternate_url"`
	Snippet     struct {
		Description  string `json:"responsibility"`
		Requirements string `json:"requirement"`
	} `json:"snippet"`
	Emptoyer struct {
		Name string `json:"name"`
	} `json:"employer"`
}

// Items ...
type Items struct {
	Vacancies []Vacancy `json:"items"`
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func main() {
	vacansies := getVacansies()
	for _, vacansy := range vacansies.Vacancies {
		if vacansy.Salary.From == 0 && vacansy.Salary.To == 0 {
			fmt.Printf("Name: %s\n", vacansy.Name)
		} else if vacansy.Salary.From != 0 && vacansy.Salary.To == 0 {
			fmt.Printf("Name: %s (от %d)\n", vacansy.Name, vacansy.Salary.From)
		} else if vacansy.Salary.From == 0 && vacansy.Salary.To != 0 {
			fmt.Printf("Name: %s (до %d)\n", vacansy.Name, vacansy.Salary.To)
		} else {
			fmt.Printf("Name: %s (%d - %d)\n", vacansy.Name, vacansy.Salary.From, vacansy.Salary.To)
		}
	}
}

// получает вакансии и мапит их в Items-структуру
func getVacansies() *Items {
	resp := sendRequest()
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	var items Items
	json.Unmarshal(body, &items)

	return &items
}

// Отправляет запрос к hh на получение вакансий
func sendRequest() *http.Response {
	url, exists := os.LookupEnv("HH_URL")
	if !exists {
		log.Fatalln("head hanter base url not found")
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
	q.Add("employment", "full") // тип занятости - полная
	q.Add("employment", "part") // или частичная
	q.Add("schedule", "remote") // тип работы - удалённая
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	return resp
}
