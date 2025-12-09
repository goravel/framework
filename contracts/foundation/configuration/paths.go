package configuration

type Paths interface {
	App(path string) Paths
	Bootstrap(path string) Paths
	Commands(path string) Paths
	Config(path string) Paths
	Controllers(path string) Paths
	Database(path string) Paths
	Events(path string) Paths
	Facades(path string) Paths
	Factories(path string) Paths
	Filters(path string) Paths
	Jobs(path string) Paths
	Lang(path string) Paths
	Listeners(path string) Paths
	Mails(path string) Paths
	Middleware(path string) Paths
	Migrations(path string) Paths
	Models(path string) Paths
	Observers(path string) Paths
	Packages(path string) Paths
	Policies(path string) Paths
	Providers(path string) Paths
	Public(path string) Paths
	Requests(path string) Paths
	Resources(path string) Paths
	Rules(path string) Paths
	Seeders(path string) Paths
	Storage(path string) Paths
	Tests(path string) Paths
}
