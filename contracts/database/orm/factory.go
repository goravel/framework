package orm

//go:generate mockery --name=Factory
type Factory interface {
	Count(count int) Factory
	Raw() any
	Create() error
	CreateQuietly() error
	Make() Factory
	Model(value any) Factory
	NewInstance(attributes ...map[string]any) Factory
}
