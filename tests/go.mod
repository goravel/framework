module github.com/goravel/framework/tests

go 1.24.0

toolchain go1.25.1

godebug x509negativeserial=1

require (
	github.com/brianvoe/gofakeit/v7 v7.6.0
	github.com/google/uuid v1.6.0
	github.com/goravel/framework v1.16.3
	github.com/goravel/mysql v1.4.0
	github.com/goravel/postgres v1.4.1
	github.com/goravel/sqlite v1.4.0
	github.com/goravel/sqlserver v1.4.0
	github.com/spf13/cast v1.10.0
	github.com/stretchr/testify v1.11.1
	gorm.io/gorm v1.31.0
)

require (
	atomicgo.dev/cursor v0.2.0 // indirect
	atomicgo.dev/keyboard v0.2.9 // indirect
	atomicgo.dev/schedule v0.1.0 // indirect
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Masterminds/squirrel v1.5.4 // indirect
	github.com/containerd/console v1.0.5 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dromara/carbon/v2 v2.6.11 // indirect
	github.com/gabriel-vasile/mimetype v1.4.10 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/gookit/color v1.5.4 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.5 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/lithammer/fuzzysearch v1.1.8 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/microsoft/go-mssqldb v1.9.2 // indirect
	github.com/ncruces/go-sqlite3 v0.26.3 // indirect
	github.com/ncruces/go-sqlite3/gormlite v0.24.0 // indirect
	github.com/ncruces/julianday v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/pterm/pterm v0.12.81 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/samber/lo v1.51.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tetratelabs/wazero v1.9.0 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	golang.org/x/crypto v0.42.0 // indirect
	golang.org/x/exp v0.0.0-20250911091902-df9299821621 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/term v0.35.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/grpc v1.75.1 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	gorm.io/driver/mysql v1.6.0 // indirect
	gorm.io/driver/postgres v1.6.0 // indirect
	gorm.io/driver/sqlserver v1.6.0 // indirect
	gorm.io/plugin/dbresolver v1.6.2 // indirect
)

replace (
	github.com/goravel/framework => ../
	github.com/goravel/mysql v0.0.0 => github.com/goravel/mysql v0.0.0
	github.com/goravel/postgres v0.0.0 => github.com/goravel/postgres v0.0.0
	github.com/goravel/sqlite v0.0.0 => github.com/goravel/sqlite v0.0.0
	github.com/goravel/sqlserver v0.0.0 => github.com/goravel/sqlserver v0.0.0
)
