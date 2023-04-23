package contracts

type Facade interface {
	GetFacadeAccessor() string
	ResolveFacadeInstance(instance any) error
}
