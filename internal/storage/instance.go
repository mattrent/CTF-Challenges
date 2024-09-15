package storage

import (
	"crypto/rand"
	"database/sql"
	"deployer/config"
	"encoding/base64"
	"time"
)

type Instance struct {
	Id          string
	ChallengeId string
	UserId      string
	Token       string
	CreatedAt   time.Time
}

func createToken(length int) (token string, err error) {
	data := make([]byte, length)
	_, err = rand.Read(data)
	if err != nil {
		return "", nil
	}
	enc := base64.URLEncoding.EncodeToString(data)
	return enc, nil
}

func CreateInstance(userId, challengeId string) (string, string, error) {
	db, err := sql.Open("postgres", config.Values.DbConn)
	if err != nil {
		return "", "", err
	}
	lastInsertId := ""
	token, err := createToken(32)
	if err != nil {
		return "", "", err
	}
	err = db.QueryRow("INSERT INTO instances (challenge_id, user_id, token) VALUES ($1, $2, $3) RETURNING id", challengeId, userId, token).Scan(&lastInsertId)
	if err != nil {
		return "", "", err
	}

	return lastInsertId, token, nil
}

func GetInstance(challengeId, token string) (Instance, error) {
	var result Instance

	db, err := sql.Open("postgres", config.Values.DbConn)
	if err != nil {
		return result, err
	}

	err = db.QueryRow("SELECT id, challenge_id, user_id, token, created_at FROM instances WHERE challenge_id = $1 AND token = $2", challengeId, token).
		Scan(&result.Id, &result.ChallengeId, &result.UserId, &result.Token, &result.CreatedAt)
	return result, err
}
