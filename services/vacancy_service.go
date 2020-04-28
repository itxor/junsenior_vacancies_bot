package services

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"parser/consts"
	"parser/db/models"
	"parser/db/repositories"
	"reflect"
	"strconv"
	"sync"
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
	var wg sync.WaitGroup
	var vacancies Items

	itemsCh := make(chan *Items, 0)
	wg.Add(1)
	go getAndUnmarshallVacancyPage(
		0,
		lastUpdateTime,
		itemsCh,
		func() { wg.Done() },
	)
	waitWgAndCloseChannel(itemsCh, &wg)
	vacancies = *(<-itemsCh)

	if vacancies.CurrentPage != 0 {
		return nil, errors.New("Невозможно распрасить страницу!")
	}

	itemsCh = make(chan *Items, vacancies.Pages)
	for i := 1; i < vacancies.Pages; i++ {
		wg.Add(1)
		go getAndUnmarshallVacancyPage(
			i,
			lastUpdateTime,
			itemsCh,
			func() { wg.Done() },
		)
	}

	waitWgAndCloseChannel(itemsCh, &wg)

	for items := range itemsCh {
		vacancies.Vacancies = append(vacancies.Vacancies, (*(items)).Vacancies...)
	}

	return &vacancies, nil
}

func waitWgAndCloseChannel(itemsCh chan *Items, wg *sync.WaitGroup) {
	go func() {
		defer close(itemsCh)
		wg.Wait()
	}()
}

func (vs *VacancyService) SaveVacancy(vacancy models.Vacancy) (bool, error) {
	return vs.vacancyRepository.InsertVacancy(vacancy)
}

func (vs *VacancyService) GetVacancyById(id int) (*models.Vacancy, error) {
	return vs.vacancyRepository.GetVacancyById(id)
}

func getAndUnmarshallVacancyPage(
	page int,
	lastUpdateTime string,
	itemsCh chan *Items,
	onExit func(),
) {
	defer onExit()

	resp, err := sendRequest(page, lastUpdateTime)
	if err != nil {
		itemsCh <- nil
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		itemsCh <- nil
		log.Fatal(err)
		return
	}

	var tempItems Items
	_ = json.Unmarshal(body, &tempItems)

	fmt.Printf("Страница %d обработана\n", page)
	itemsCh <- mergeVacancies(&tempItems)
	fmt.Printf("Данные записаны в канал\n")

	return
}

// Отправляет запрос к hh на получение вакансий
func sendRequest(page int, dateFrom string) (*http.Response, error) {
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

	q.Add("text", "NAME:(PHP OR Symfony OR Laravel OR Backend OR Back-end OR BackEnd) "+
		" AND DESCRIPTION:(PHP OR php) "+
		"NOT Bitrix NOT BITRIX NOT 1C NOT 1С NOT 1c ")
	q.Add("employment", "full")       // тип занятости - полная
	q.Add("employment", "part")       // или частичная
	q.Add("schedule", "remote")       // тип работы - удалённая
	q.Add("page", strconv.Itoa(page)) // страница
	q.Add("specialization", strconv.Itoa( // профобласть
		consts.PROGRAMMING_SPECIALIZATION_ID,
	))
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

func mergeVacancies(items *Items) *Items {
	vacancies := items.Vacancies

	excludedKeys := make([]int, 0)
	doubles := make(map[int][]int)
	withoutDoubleVacancies := make([]models.Vacancy, 0)

	for i := 0; i < len(vacancies); i++ {
		if itemExists(excludedKeys, i) {
			continue
		} else {
			excludedKeys = append(excludedKeys, i)
		}

		for j := 0; j < len(vacancies); j++ {
			if itemExists(excludedKeys, j) {
				continue
			}

			if isDouble(vacancies, i, j) {
				excludedKeys = append(excludedKeys, j)
				if len(doubles[i]) == 0 {
					doubles[i] = make([]int, 0)
				}
				doubles[i] = append(doubles[i], j)
				continue
			}
		}

		if _, ok := doubles[i]; !ok {
			withoutDoubleVacancies = append(withoutDoubleVacancies, vacancies[i])
		}
	}

	if len(doubles) != 0 {

		for uniqueIndex, doublesIndexes := range doubles {
			mergedPlace := vacancies[uniqueIndex].Area.Place + ", "
			for index, doublesIndex := range doublesIndexes {
				mergedPlace += vacancies[doublesIndex].Area.Place
				if index != len(doublesIndexes)-1 {
					mergedPlace += ", "
				}
			}

			vacancies[uniqueIndex].Area.Place = mergedPlace
			withoutDoubleVacancies = append(withoutDoubleVacancies, vacancies[uniqueIndex])
		}

		items.Vacancies = withoutDoubleVacancies
	}

	return items
}

func isDouble(vacancies []models.Vacancy, i int, j int) bool {
	return vacancies[i].Name == vacancies[j].Name &&
		vacancies[i].Salary.From == vacancies[j].Salary.From &&
		vacancies[i].Salary.To == vacancies[j].Salary.To &&
		vacancies[i].Salary.Currency == vacancies[j].Salary.Currency &&
		vacancies[i].Employer.Name == vacancies[j].Employer.Name
}

func itemExists(slice interface{}, item interface{}) bool {
	s := reflect.ValueOf(slice)

	if s.Kind() != reflect.Slice {
		panic("Invalid data-type")
	}

	for i := 0; i < s.Len(); i++ {
		if s.Index(i).Interface() == item {
			return true
		}
	}

	return false
}
