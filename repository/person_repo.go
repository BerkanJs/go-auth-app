package repository

import (
	"database/sql"

	"go-kisi-api/db"
	"go-kisi-api/models"
	"go-kisi-api/queries"
)

func AddPerson(p models.Person) (int64, error) {
	result, err := db.DB.Exec(
		queries.InsertPerson,
		p.Name, p.Surname, p.Email, p.Age, p.Phone, p.PasswordHash,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetAllPeople() ([]models.Person, error) {
	rows, err := db.DB.Query(queries.SelectAllPeople)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []models.Person
	for rows.Next() {
		var p models.Person
		if err := rows.Scan(&p.ID, &p.Name, &p.Surname, &p.Email, &p.Age, &p.Phone, &p.PasswordHash); err != nil {
			return nil, err
		}
		people = append(people, p)
	}
	return people, nil
}

func GetPersonByID(id int) (models.Person, error) {
	var p models.Person
	row := db.DB.QueryRow(queries.SelectPersonByID, id)
	err := row.Scan(&p.ID, &p.Name, &p.Surname, &p.Email, &p.Age, &p.Phone, &p.PasswordHash)
	return p, err
}

func GetPersonByEmail(email string) (models.Person, error) {
	var p models.Person
	row := db.DB.QueryRow(queries.SelectPersonByEmail, email)
	err := row.Scan(&p.ID, &p.Name, &p.Surname, &p.Email, &p.Age, &p.Phone, &p.PasswordHash)
	return p, err
}

// EmailExists verilen email için bir kayıt olup olmadığını döner.
func EmailExists(email string) (bool, error) {
	var id int
	err := db.DB.QueryRow(queries.SelectPersonIDByEmail, email).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func DeletePerson(id int) error {
	_, err := db.DB.Exec(queries.DeletePersonByID, id)
	return err
}
