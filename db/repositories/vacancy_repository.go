package repositories

import (
	"database/sql"
	"log"
	"os"
	"parser/db/models"
)

type VacancyRepository struct {
	databaseURL    string
	databaseDriver string
}

func InitDB() (*VacancyRepository, error) {
	databaseUrl, exists := os.LookupEnv("DATABASE_URL")
	if !exists {
		log.Fatalln("database url is not set")
	}
	databaseDriver, exists := os.LookupEnv("DATABASE_DRIVER")
	if !exists {
		log.Fatalln("database driver is not set")
	}

	return &VacancyRepository{
		databaseURL:    databaseUrl,
		databaseDriver: databaseDriver,
	}, nil
}

// InsertVacancy ... 
func (database *VacancyRepository) InsertVacancy(vacancy models.Vacancy) (bool, error) {
	db, err := sql.Open(
		database.databaseDriver,
		database.databaseURL,
	)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO vacancies( " +
		"id, name, place, salary_from, salary_to, salary_currency, salary_gross, published_at, archived, " +
		"url, employer_name) " +
		"SELECT ?,?,?,?,?,?,?,?,?,?,? " +
		"WHERE NOT EXISTS (SELECT * FROM vacancies WHERE id = (?))")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(
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
		vacancy.ID, // для поиска уже существующего значения
	)
	if err != nil {
		return false, err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	if count != 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func (database *VacancyRepository) GetVacancyById(id int) (*models.Vacancy, error) {
	db, err := sql.Open(
		database.databaseDriver,
		database.databaseURL,
	)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(
		"SELECT id, "+
			"name, "+
			"place, "+
			"salary_from, "+
			"salary_to, "+
			"salary_currency, "+
			"salary_gross, "+
			"published_at, "+
			"archived, "+
			"url, "+
			"employer_name "+
			" FROM vacancies WHERE id=?",
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rows.Next()

	vacancy := &models.Vacancy{}
	if err = rows.Scan(
		&vacancy.ID,
		&vacancy.Name,
		&vacancy.Area.Place,
		&vacancy.Salary.From,
		&vacancy.Salary.To,
		&vacancy.Salary.Currency,
		&vacancy.Salary.Gross,
		&vacancy.PublishedAt,
		&vacancy.Archived,
		&vacancy.URL,
		&vacancy.Employer.Name,
	); err != nil {
		log.Fatal(err)

		return nil, err
	}

	return vacancy, nil
}
