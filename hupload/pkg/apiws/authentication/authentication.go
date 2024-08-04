package authentication

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// AuthenticationInterface must be implemented by the authentication backend
type AuthenticationInterface interface {
	AuthenticateUser(username, password string) (bool, error)
}
