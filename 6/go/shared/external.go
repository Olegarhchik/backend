package shared

import (
	"database/sql"
	"fmt"
	"strings"
)

type Application struct {
	Login string
	FullName string
	Phone string
	Email string
	Birthdate string
	Gender string
	ProgLang []string
	Bio string
}

func GetUser(id string) (Application, error) {
	info := Application{}

	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return info, err
	}

	defer db.Close()

	sel, err := db.Query(fmt.Sprintf(`
		SELECT FullName, PhoneNumber, APPL.Email, Birthdate, Gender, Biography
		FROM Application APPL
		JOIN User U ON ApplicationID = Login
		WHERE APPL.ApplicationID = '%s';
	`, id))

	if err != nil {
		return info, err
	}

	defer sel.Close()

	for sel.Next() {
		err := sel.Scan(&info.FullName, &info.Phone, &info.Email, &info.Birthdate, &info.Gender, &info.Bio)

		if err != nil {
			return info, err
		}
	}

	sel, err = db.Query(fmt.Sprintf(`
		SELECT Name
		FROM ProgLang PL
		JOIN Abilities A
		ON PL.ProgLangID = A.ProgLangID
		WHERE A.ApplicationID = '%s';
	`, id))

	if err != nil {
		return info, err
	}

	defer sel.Close()

	for sel.Next() {
		var pl string
		err := sel.Scan(&pl)

		if err != nil {
			return info, err
		}

		info.ProgLang = append(info.ProgLang, pl)
	}

	return info, nil
}

func UpdateCol(column string, newValue string, IDName string, ID string, table string) error {
	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return err
	}

	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(`
		UPDATE %s
		SET %s = ?
		WHERE %s = ?;
	`, table, column, IDName), newValue, ID)

	if err != nil {
		return err
	}

	return nil
}

func Contains(array []string, elem string) bool {
	for _, v := range array {
		if v == elem {
			return true
		}
	}

	return false
}

func UpdateCols(old Application, new Application, login string) error {
	if old.FullName != new.FullName {
		err := UpdateCol("FullName", new.FullName, "ApplicationID", login, "Application")

		if err != nil {
			return err
		}
	}

	if old.Phone != new.Phone {
		err := UpdateCol("PhoneNumber", new.Phone, "ApplicationID", login, "Application")

		if err != nil {
			return err
		}
	}

	if old.Birthdate != new.Birthdate {
		err := UpdateCol("Birthdate", new.Birthdate, "ApplicationID", login, "Application")

		if err != nil {
			return err
		}
	}

	if old.Gender != new.Gender {
		err := UpdateCol("Gender", new.Gender, "ApplicationID", login, "Application")

		if err != nil {
			return err
		}
	}

	if old.Bio != new.Bio {
		err := UpdateCol("Biography", new.Bio, "ApplicationID", login, "Application")

		if err != nil {
			return err
		}
	}

	for _, pl := range new.ProgLang {
		if (!Contains(old.ProgLang, pl)) {
			err := InsertPL(login, pl)

			if err != nil {
				return err
			}
		}
	}

	for _, pl := range old.ProgLang {
		if (!Contains(new.ProgLang, pl)) {
			err := deletePL(login, pl)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func InsertPL(applID string, PLName string) error {
	PLID, err := getPLID(PLName)

	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return err
	}

	defer db.Close()

	return InsValues("Abilities", "ApplicationID, ProgLangID", applID, PLID)
}

func InsValues(table string, cols string, values ...string) error {
	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return err
	}

	defer db.Close()

	for i, value := range values {
		values[i] = "'" + value + "'"
	}

	_, err = db.Exec(fmt.Sprintf(`
		INSERT INTO %s(%s)
		VALUES (%s)
	`, table, cols, strings.Join(values, ", ")))

	if err != nil {
		return err
	}

	return nil
}