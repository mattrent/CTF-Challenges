package storage

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

type Instance struct {
	Id          string    `json:"id"`
	ChallengeId string    `json:"challenge_id"`
	PlayerId    string    `json:"player_id"`
	Token       string    `json:"token"`
	CreatedAt   time.Time `json:"created_at"`
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
	lastInsertId := ""
	token, err := createToken(32)
	if err != nil {
		return "", "", err
	}
	err = Db.QueryRow("INSERT INTO instances (challenge_id, player_id, token) VALUES ($1, $2, $3) RETURNING id", challengeId, userId, token).Scan(&lastInsertId)
	if err != nil {
		return "", "", err
	}

	return lastInsertId, token, nil
}

func GetInstance(challengeId, token string) (Instance, error) {
	var result Instance

	err := Db.QueryRow("SELECT id, challenge_id, player_id, token, created_at FROM instances WHERE challenge_id = $1 AND token = $2", challengeId, token).
		Scan(&result.Id, &result.ChallengeId, &result.PlayerId, &result.Token, &result.CreatedAt)
	return result, err
}
