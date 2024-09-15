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
