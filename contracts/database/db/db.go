package db

type DB interface {
	Table(name string) Query
}

type Query interface {
	Where(query any, args ...any) Query
	Get(dest any) error
}
