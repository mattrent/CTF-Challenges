package storage

type User struct {
	Id           string
	Username     string
	PasswordHash string
	Role         string
}

func GetUser(username string) (User, error) {
	var user User

	err := Db.QueryRow("SELECT id, username, password_hash, role FROM users WHERE username=$1;", username).Scan(&user.Id, &user.Username, &user.PasswordHash, &user.Role)
	if err != nil {
		return user, err
	}

	return user, nil
}

func CreateUser(username, passwordHash, role string) error {
	_, err := Db.Exec("INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3)", username, passwordHash, role)
	if err != nil {
		return err
	}

	return nil
}
