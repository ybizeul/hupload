package authentication

import (
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
)

// AuthenticationFile takes users from a yaml file
// Example :
// - username: admin
//   password: $2y$10$ro2aBKU9jyqfokF2arnaEO3GKmAawnfLfEFq1dGuGl9CYEutrxGCa
// - username: test
//   password: $2y$10$ro2aBKU9jyqfokF2arnaEO3GKmAawnfLfEFq1dGuGl9CYEutrxGCa

type FileAuthenticationConfig struct {
	Path string `yaml:"path"`
}

type AuthenticationFile struct {
	Options FileAuthenticationConfig
}

func NewAuthenticationFile(o FileAuthenticationConfig) (*AuthenticationFile, error) {
	r := AuthenticationFile{
		Options: o,
	}

	path := r.Options.Path

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrAuthenticationMissingUsersFile
		}
		return nil, err
	}
	return &r, nil
}

func (a *AuthenticationFile) AuthenticateRequest(w http.ResponseWriter, r *http.Request) error {
	username, password, ok := r.BasicAuth()
	if !ok {
		return ErrAuthenticationMissingCredentials
	}

	// Prepare struct to load users.yaml
	var users []User

	path := a.Options.Path

	// Fail if we can't open the file
	pf, err := os.Open(path)
	if err != nil {
		return err
	}
	defer pf.Close()

	// Load users.yml
	err = yaml.NewDecoder(pf).Decode(&users)
	if err != nil {
		return err
	}

	// Check if user is in the list
	for _, u := range users {
		if u.Username == username {
			// Compare password hash
			err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
			if err == nil {
				return nil
			}
		}
	}

	return ErrAuthenticationBadCredentials
}

func (o *AuthenticationFile) CallbackFunc(http.Handler) (func(w http.ResponseWriter, r *http.Request), bool) {
	return nil, false
}

func (o *AuthenticationFile) ShowLoginForm() bool {
	return true
}

func (o *AuthenticationFile) LoginURL() string {
	return "/"
}
