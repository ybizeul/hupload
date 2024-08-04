package authentication

import (
	"os"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v2"
)

type AuthBackendBasic struct {
	Options map[string]any
}

func NewAuthenticationBasic(m map[string]any) *AuthBackendBasic {
	r := &AuthBackendBasic{
		Options: m["options"].(map[string]any),
	}

	return r
}
func (a *AuthBackendBasic) AuthenticateUser(username, password string) (bool, error) {
	// Prepare struct to load users.yaml
	users := []User{}

	path := a.Options["path"].(string)

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
			if err != nil {
				return false, nil
			}
			return true, nil
		}
	}
	return false, nil
}
