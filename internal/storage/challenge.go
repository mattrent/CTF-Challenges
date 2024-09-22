package storage

import (
	"database/sql"
	"deployer/config"
)

type Challenge struct {
	Id        string
	UserId    string
	Published bool
}

func GetChallenge(challengeId string) (Challenge, error) {
	var result Challenge

	db, err := sql.Open("postgres", config.Values.DbConn)
	if err != nil {
		return result, err
	}

	err = db.QueryRow("SELECT id, user_id, published FROM challenges WHERE id=$1;", challengeId).Scan(&result.Id, &result.UserId, &result.Published)
	if err != nil {
		return result, err
	}

	return result, nil
}

func ListChallenges() ([]Challenge, error) {
	var result []Challenge

	db, err := sql.Open("postgres", config.Values.DbConn)
	if err != nil {
		return result, err
	}

	rows, err := db.Query("SELECT id, user_id, published FROM challenges;")
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var challenge Challenge
		err := rows.Scan(&challenge.Id, &challenge.UserId, &challenge.Published)
		if err != nil {
			return result, err
		}
		result = append(result, challenge)
	}
	if err := rows.Err(); err != nil {
		return result, err
	}
	return result, nil
}

func CreateChallenge(userId string) (string, error) {
	db, err := sql.Open("postgres", config.Values.DbConn)
	if err != nil {
		return "", err
	}
	lastInsertId := ""
	err = db.QueryRow("INSERT INTO challenges (user_id) VALUES ($1) RETURNING id", userId).Scan(&lastInsertId)
	if err != nil {
		return lastInsertId, err
	}

	return lastInsertId, nil
}

func PublishChallenge(challengeId string) error {
	db, err := sql.Open("postgres", config.Values.DbConn)
	if err != nil {
		return err
	}
	_, err = db.Exec("UPDATE challenges SET published=$1 WHERE id=$2", true, challengeId)
	if err != nil {
		return err
	}
	return nil
}

func DeleteChallenge(challengeId string) error {
	db, err := sql.Open("postgres", config.Values.DbConn)
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM challenges WHERE id=$1", challengeId)
	if err != nil {
		return err
	}
	return nil
}
