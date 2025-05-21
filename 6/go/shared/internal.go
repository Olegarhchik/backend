package shared

import (
	"database/sql"
	"fmt"
)

func deletePL(applID string, PLName string) error {
	PLID, err := getPLID(PLName)

	if err != nil {
		return err
	}

	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf(`
		DELETE FROM Abilities
		WHERE ApplicationID = '%s' AND ProgLangID = '%s';
	`, applID, PLID))

	if err != nil {
		return err
	}

	return nil
}

func getPLID(PLName string) (string, error) {
	PLID := ""

	db, err := sql.Open("mysql", "u68861:1067131@/u68861")

	if err != nil {
		return PLID, err
	}

	defer db.Close()

	sel, err := db.Query(fmt.Sprintf(`
		SELECT ProgLangID
		FROM ProgLang
		WHERE Name = '%s';
	`, PLName))

	if err != nil {
		return PLID, err
	}

	defer sel.Close()

	for sel.Next() {
		err := sel.Scan(&PLID)

		if err != nil {
    		return PLID, err
  		}
	}

	return PLID, nil
}