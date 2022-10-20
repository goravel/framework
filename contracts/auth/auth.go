package auth

//go:generate mockery --name=Auth
type Auth interface {
	Guard(name string) Auth
	Parse(token string) (expired bool, err error)
	User(user interface{}) error
	Login(user interface{}) (token string, err error)
	LoginUsingID(id interface{}) (token string, err error)
	Refresh() (token string, err error)
	Logout() error
}
