package orm

//go:generate mockery --name=Factory
type Factory interface {
	Times(count int) Factory
	Create(value any) error
	CreateQuietly(value any) error
	Make(value any) error
}
