package authentication

import (
	"errors"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v2"
)

// AuthenticationFile takes users from a yaml file
// Example :
// - username: admin
//   password: $2y$10$ro2aBKU9jyqfokF2arnaEO3GKmAawnfLfEFq1dGuGl9CYEutrxGCa
// - username: test
//   password: $2y$10$ro2aBKU9jyqfokF2arnaEO3GKmAawnfLfEFq1dGuGl9CYEutrxGCa

type AuthenticationFile struct {
	Options map[string]any
}

func NewAuthenticationFile(m map[string]any) *AuthenticationFile {
	r := &AuthenticationFile{
		Options: m["options"].(map[string]any),
	}

	return r
}

func (a *AuthenticationFile) AuthenticateUser(username, password string) (bool, error) {
	// Prepare struct to load users.yaml
	users := []User{}

	path, ok := a.Options["path"].(string)

	// Fail is cast to string didn't work
	if !ok {
		return false, errors.New("path option is invalid")
	}

	// Fail if we can't open the file
	pf, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer pf.Close()

	// Load users.yml
	err = yaml.NewDecoder(pf).Decode(&users)
	if err != nil {
		return false, err
	}

	// Check if user is in the list
	for _, u := range users {
		if u.Username == username {
			// Compare password hash
			err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
			if err == nil {
				return true, nil
			}
		}
	}
	return false, nil
}
