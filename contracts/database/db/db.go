package db

type DB interface {
	Table(name string) Query
}

type Query interface {
	Delete() error
	Get(dest any) error
	Insert() error
	Update() error
	Where(query any, args ...any) Query
}
