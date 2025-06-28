module github.com/goravel/framework/tests

go 1.23.0

toolchain go1.24.4

godebug x509negativeserial=1

require (
	github.com/brianvoe/gofakeit/v7 v7.2.1
	github.com/goravel/framework v1.15.9
	github.com/goravel/mysql v0.0.0-20250626031516-19442c544454
	github.com/goravel/postgres v0.0.2-0.20250626031535-aae7c8a0767e
	github.com/goravel/sqlite v0.0.0-20250626031454-e195ad35c34e
	github.com/goravel/sqlserver v0.0.0-20250627011359-b501c1567c71
	github.com/spf13/cast v1.9.2
	github.com/stretchr/testify v1.10.0
	gorm.io/gorm v1.30.0
)

require (
	atomicgo.dev/cursor v0.2.0 // indirect
	atomicgo.dev/keyboard v0.2.9 // indirect
	atomicgo.dev/schedule v0.1.0 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Masterminds/semver/v3 v3.3.1 // indirect
	github.com/Masterminds/squirrel v1.5.4 // indirect
	github.com/containerd/console v1.0.5 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dromara/carbon/v2 v2.6.9 // indirect
	github.com/gabriel-vasile/mimetype v1.4.9 // indirect
	github.com/go-sql-driver/mysql v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.3.0 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gookit/color v1.5.4 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.2 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lithammer/fuzzysearch v1.1.8 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/microsoft/go-mssqldb v1.9.1 // indirect
	github.com/ncruces/go-sqlite3 v0.25.0 // indirect
	github.com/ncruces/go-sqlite3/gormlite v0.24.0 // indirect
	github.com/ncruces/julianday v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/pterm/pterm v0.12.81 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/samber/lo v1.51.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tetratelabs/wazero v1.9.0 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250528174236-200df99c418a // indirect
	google.golang.org/grpc v1.73.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.6.0 // indirect
	gorm.io/driver/postgres v1.6.0 // indirect
	gorm.io/driver/sqlserver v1.6.0 // indirect
	gorm.io/plugin/dbresolver v1.6.0 // indirect
)

replace (
	github.com/goravel/framework => ../
	github.com/goravel/mysql v0.0.0 => github.com/goravel/mysql v0.0.0
	github.com/goravel/postgres v0.0.0 => github.com/goravel/postgres v0.0.0
	github.com/goravel/sqlite v0.0.0 => github.com/goravel/sqlite v0.0.0
	github.com/goravel/sqlserver v0.0.0 => github.com/goravel/sqlserver v0.0.0
)
