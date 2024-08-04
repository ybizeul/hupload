package authentication

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// BackendInterface must be implemented by any backend
type AuthenticationInterface interface {
	AuthenticateUser(username, password string) (bool, error)
}
