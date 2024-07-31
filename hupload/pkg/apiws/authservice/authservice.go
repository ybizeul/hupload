package authservice

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// BackendInterface must be implemented by any backend
type AuthServiceInterface interface {
	AuthenticateUser(username, password string) (bool, error)
}
