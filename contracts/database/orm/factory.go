package orm

//go:generate mockery --name=Factory
type Factory interface {
	Count(count int) Factory
	Create(value any, attributes ...map[string]any) error
	CreateQuietly(value any, attributes ...map[string]any) error
	Make(value any, attributes ...map[string]any) error
}
