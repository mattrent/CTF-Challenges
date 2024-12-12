package storage

import (
	"database/sql"
)

type Challenge struct {
	Id        string        `json:"id"`
	UserId    string        `json:"user_id"`
	Published bool          `json:"published"`
	CtfdId    sql.NullInt64 `json:"ctfd_id"`
}

func GetChallenge(challengeId string) (Challenge, error) {
	var result Challenge

	err := Db.QueryRow("SELECT id, user_id, published, ctfd_id FROM challenges WHERE id=$1;", challengeId).Scan(&result.Id, &result.UserId, &result.Published, &result.CtfdId)
	if err != nil {
		return result, err
	}

	return result, nil
}

func GetChallengeByCtfdId(ctfdId int) (Challenge, error) {
	var result Challenge

	err := Db.QueryRow("SELECT id, user_id, published, ctfd_id FROM challenges WHERE ctfd_id=$1;", ctfdId).Scan(&result.Id, &result.UserId, &result.Published, &result.CtfdId)
	if err != nil {
		return result, err
	}

	return result, nil
}

func ListChallenges() ([]Challenge, error) {
	var result []Challenge

	rows, err := Db.Query("SELECT id, user_id, published, ctfd_id FROM challenges;")
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var challenge Challenge
		err := rows.Scan(&challenge.Id, &challenge.UserId, &challenge.Published, &challenge.CtfdId)
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
	if err != nil {
		return lastInsertId, err
	}

	return lastInsertId, nil
}

func PublishChallengeWithReference(challengeId string, ctfdId int) error {
	_, err := Db.Exec("UPDATE challenges SET published=$1, ctfd_id=$2 WHERE id=$3", true, ctfdId, challengeId)
	if err != nil {
		return err
	}
	return nil
}

func DeleteChallenge(challengeId string) error {
	_, err := Db.Exec("DELETE FROM challenges WHERE id=$1", challengeId)
	if err != nil {
		return err
	}
	return nil
}
