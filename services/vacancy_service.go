package services

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"parser/db/models"
	"parser/db/repositories"
	"strconv"
)

// Items ...
type Items struct {
	Vacancies   []models.Vacancy `json:"items"`
	Pages       int              `json:"pages"`
	CurrentPage int              `json:"page"`
}

type VacancyService struct {
	vacancyRepository *repositories.VacancyRepository
}

func CreateVacancyService() (*VacancyService, error) {
	vacancyRepository, err := repositories.InitDB()
	if err != nil {
		return nil, err
	}

	return &VacancyService{
		vacancyRepository: vacancyRepository,
	}, nil
}

// получает вакансии и мапит их в Items-структуру
func (vs *VacancyService) GetVacancies(lastUpdateTime string) (*Items, error) {
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

// SaveVacancy ...
func (vs *VacancyService) SaveVacancy(vacancy models.Vacancy) (bool, error) {
	return vs.vacancyRepository.InsertVacancy(vacancy)
}

// Отправляет запрос к hh на получение вакансий
func sendRequest(page int, dateFrom string) (*http.Response, error) {
	url, exists := os.LookupEnv("HH_URL")
	if !exists {
		log.Fatalln("head hanter base url not found")

		return nil, errors.New("URL HH не найдено!")
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
