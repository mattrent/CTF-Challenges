package storage

import (
	"database/sql"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Challenge struct {
	Id        string        `json:"id"`
	UserId    string        `json:"user_id"`
	Published bool          `json:"published"`
	CtfdId    sql.NullInt64 `json:"ctfd_id"`
	Verified  bool          `json:"verified"`
}

type Config struct {
	Name        string   `yaml:"name"`
	Author      string   `yaml:"author"`
	Category    string   `yaml:"category"`
	Description string   `yaml:"description"`
	Type        string   `yaml:"type"`
	Extra       Extra    `yaml:"extra"`
	Solution    string   `yaml:"solution"`
	Flags       []string `yaml:"flags"`
}

// Define a struct for the "extra" field
type Extra struct {
	Function string `yaml:"function"`
	Initial  int    `yaml:"initial"`
	Decay    int    `yaml:"decay"`
	Minimum  int    `yaml:"minimum"`
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

func ParseChallengeYAML(filePath string) (Config, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return Config{}, fmt.Errorf("error reading file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(fileContent, &config)
	if err != nil {
		return Config{}, fmt.Errorf("error parsing YAML: %v", err)
	}

	return config, nil
}

func UpdateChallengeFlagGivenChallengeFile(filePath string, challengeId string) error {
	config, err := ParseChallengeYAML(filePath)
	if err != nil {
		return err
	}
	err = UpdateChallengeFlag(challengeId, config.Flags[0])
	return err
}

func ListChallenges() ([]Challenge, error) {
	var result []Challenge

	rows, err := Db.Query("SELECT id, user_id, published, ctfd_id, verified FROM challenges;")
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
