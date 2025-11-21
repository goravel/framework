package configuration

type Paths interface {
	Bootstrap(path string) Paths
	Command(path string) Paths
	Controller(path string) Paths
	Event(path string) Paths
	Factory(path string) Paths
	Filter(path string) Paths
	Job(path string) Paths
	Listener(path string) Paths
	Mail(path string) Paths
	Middleware(path string) Paths
	Migration(path string) Paths
	Model(path string) Paths
	Observer(path string) Paths
	Package(path string) Paths
	Policy(path string) Paths
	Provider(path string) Paths
	Request(path string) Paths
	Rule(path string) Paths
	Seeder(path string) Paths
	Test(path string) Paths
}
