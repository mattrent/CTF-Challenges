package storage

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

type Challenge struct {
	Id        string        `json:"id"`
	UserId    string        `json:"user_id"`
	Published bool          `json:"published"`
	CtfdId    sql.NullInt64 `json:"ctfd_id"`
	Verified  bool          `json:"verified"`
}

type ChallengeCtfd struct {
	Name           string        `json:"name"`
	Author         string        `json:"author"`
	Category       string        `json:"category"`
	Description    string        `json:"description"`
	Value          int           `json:"value"`
	Type           string        `json:"type"`
	Extra          *Extra        `json:"extra"`
	Image          interface{}   `json:"image"`
	Protocol       interface{}   `json:"protocol"`
	Host           interface{}   `json:"host"`
	ConnectionInfo string        `json:"connection_info"`
	Healthcheck    string        `json:"healthcheck"`
	Attempts       int           `json:"attempts"`
	Flags          []string      `json:"flags"`
	Topics         []string      `json:"topics"`
	Tags           []string      `json:"tags"`
	Files          []string      `json:"files"`
	Hints          []HintElement `json:"hints"`
	Requirements   []string      `json:"requirements"`
	State          string        `json:"state"`
	Version        string        `json:"version"`
}

type Extra struct {
	Initial  int    `json:"initial"`
	Decay    int    `json:"decay"`
	Minimum  int    `json:"minimum"`
	Function string `json:"function"`
}

type FlagClass struct {
	Type    string  `json:"type"`
	Content string  `json:"content"`
	Data    *string `json:"data,omitempty"`
}

type HintClass struct {
	Content string `json:"content"`
	Cost    int    `json:"cost"`
}

type HintElement struct {
	HintClass *HintClass
	String    *string
}

func GetChallenge(challengeId string) (Challenge, error) {
	var result Challenge

	err := Db.QueryRow("SELECT id, user_id, published, ctfd_id, verified FROM challenges WHERE id=$1;", challengeId).Scan(&result.Id, &result.UserId, &result.Published, &result.CtfdId, &result.Verified)
	return result, err
}

func GetChallengeByCtfdId(ctfdId int) (Challenge, error) {
	var result Challenge

	err := Db.QueryRow("SELECT id, user_id, published, ctfd_id, verified FROM challenges WHERE ctfd_id=$1;", ctfdId).Scan(&result.Id, &result.UserId, &result.Published, &result.CtfdId, &result.Verified)
	return result, err
}

func UpdateChallengeFlag(challengeId string, flag string) error {
	_, err := Db.Exec("UPDATE challenges SET flag=$1 WHERE id=$2", flag, challengeId)
	return err
}

func GetChallengeWrapper(challengeId string) (Challenge, error) {
	var challenge Challenge
	if id, err := strconv.Atoi(challengeId); err == nil {
		challenge, err = GetChallengeByCtfdId(id)
		if err != nil {
			return Challenge{}, fmt.Errorf("CTFd challenge not found")
		}
	} else {
		challenge, err = GetChallenge(challengeId)
		if err != nil {
			return Challenge{}, fmt.Errorf("Challenge not found")
		}
	}

	return challenge, nil
}

func GetChallengeFlag(challengeId string) (string, error) {
	var expectedFlag string
	err := Db.QueryRow("SELECT flag FROM challenges WHERE id=$1;", challengeId).Scan(&expectedFlag)
	if err != nil {
		return "", err
	}
	return expectedFlag, err
}

func MarkChallengeVerified(challengeId string) error {
	_, err := Db.Exec("UPDATE challenges SET verified=$1 WHERE id=$2", true, challengeId)
	return err
}

func ResetChallengeVerified(challengeId string) error {
	_, err := Db.Exec("UPDATE challenges SET verified=$1 WHERE id=$2", false, challengeId)
	return err
}

func ParseChallengeYAML(filePath string) (ChallengeCtfd, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return ChallengeCtfd{}, fmt.Errorf("error reading file: %v", err)
	}

	var config ChallengeCtfd
	err = yaml.Unmarshal(fileContent, &config)
	if err != nil {
		return ChallengeCtfd{}, fmt.Errorf("error parsing YAML: %v", err)
	}

	return config, nil
}

// ! Only support challenges with a single flag
func UpdateChallengeFlagGivenChallengeFile(filePath string, challengeId string) error {
	config, err := ParseChallengeYAML(filePath)
	if err != nil {
		return err
	}
	err = UpdateChallengeFlag(challengeId, config.Flags[0])
	return err
}

func ListChallenges(isAdmin bool, userId string) ([]Challenge, error) {
	var result []Challenge
	var rows *sql.Rows
	var err error

	if isAdmin {
		rows, err = Db.Query(
			"SELECT id, user_id, published, ctfd_id, verified FROM challenges;",
		)
	} else {
		rows, err = Db.Query(
			"SELECT id, user_id, published, ctfd_id, verified FROM challenges WHERE user_id = $1;",
			userId,
		)
	}

	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var challenge Challenge
		err := rows.Scan(&challenge.Id, &challenge.UserId, &challenge.Published, &challenge.CtfdId, &challenge.Verified)
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
	lastInsertId := ""
	err := Db.QueryRow("INSERT INTO challenges (user_id) VALUES ($1) RETURNING id", userId).Scan(&lastInsertId)
	return lastInsertId, err
}

func PublishChallengeWithReference(challengeId string, ctfdId int) error {
	_, err := Db.Exec("UPDATE challenges SET published=$1, ctfd_id=$2 WHERE id=$3", true, ctfdId, challengeId)
	return err
}

func DeleteChallenge(challengeId string) error {
	_, err := Db.Exec("DELETE FROM challenges WHERE id=$1", challengeId)
	return err
}
