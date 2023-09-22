package translation

//go:generate mockery --name=Loader
type Loader interface {
	Load(folder string, locale string) (map[string]interface{}, error)
}
