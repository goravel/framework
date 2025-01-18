package tests

import (
	contractsdatabase "github.com/goravel/framework/contracts/database"
)

type TestTable int

const (
	TestTableAddresses TestTable = iota
	TestTableAuthors
	TestTableBooks
	TestTableHouses
	TestTablePeoples
	TestTablePhones
	TestTableProducts
	TestTableReviews
	TestTableRoles
	TestTableRoleUser
	TestTableUsers
	TestTableGoravelUser
	TestTableSchema
)

type testTables struct {
	driver contractsdatabase.Driver
}

func newTestTables(driver contractsdatabase.Driver) *testTables {
	return &testTables{driver: driver}
}

func (r *testTables) All() map[TestTable]func() string {
	return map[TestTable]func() string{
		TestTableAddresses:   r.addresses,
		TestTableAuthors:     r.authors,
		TestTableBooks:       r.books,
		TestTableHouses:      r.houses,
		TestTablePeoples:     r.peoples,
		TestTablePhones:      r.phones,
		TestTableProducts:    r.products,
		TestTableReviews:     r.reviews,
		TestTableRoles:       r.roles,
		TestTableRoleUser:    r.roleUser,
		TestTableUsers:       r.users,
		TestTableGoravelUser: r.goravelUser,
		TestTableSchema:      r.schema,
	}
}

func (r *testTables) peoples() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE peoples (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  body varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE peoples (
  id SERIAL PRIMARY KEY NOT NULL,
  body varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE peoples (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  body varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE peoples (
  id bigint NOT NULL IDENTITY(1,1),
  body varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) reviews() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE reviews (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  body varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE reviews (
  id SERIAL PRIMARY KEY NOT NULL,
  body varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE reviews (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  body varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE reviews (
  id bigint NOT NULL IDENTITY(1,1),
  body varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) products() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE products (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE products (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE products (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE products (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) users() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE users (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE users (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE users (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE users (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) goravelUser() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE goravel_user (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  deleted_at datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_users_created_at (created_at),
  KEY idx_users_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE goravel_user (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  deleted_at timestamp DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE goravel_user (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE goravel_user (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  bio varchar(255) DEFAULT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  deleted_at datetime DEFAULT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) addresses() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE addresses (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  user_id bigint(20) unsigned DEFAULT NULL,
  name varchar(255) NOT NULL,
  province varchar(255) DEFAULT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_addresses_created_at (created_at),
  KEY idx_addresses_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE addresses (
  id SERIAL PRIMARY KEY NOT NULL,
  user_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  province varchar(255) DEFAULT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE addresses (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  user_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  province varchar(255) DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE addresses (
  id bigint NOT NULL IDENTITY(1,1),
  user_id bigint DEFAULT NULL,
  name varchar(255) NOT NULL,
  province varchar(255) DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) books() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE books (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  user_id bigint(20) unsigned DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_books_created_at (created_at),
  KEY idx_books_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE books (
  id SERIAL PRIMARY KEY NOT NULL,
  user_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE books (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  user_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE books (
  id bigint NOT NULL IDENTITY(1,1),
  user_id bigint DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) authors() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE authors (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  book_id bigint(20) unsigned DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_books_created_at (created_at),
  KEY idx_books_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE authors (
  id SERIAL PRIMARY KEY NOT NULL,
  book_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE authors (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  book_id int DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE authors (
  id bigint NOT NULL IDENTITY(1,1),
  book_id bigint DEFAULT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) roles() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE roles (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_roles_created_at (created_at),
  KEY idx_roles_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE roles (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE roles (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE roles (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  avatar varchar(255) DEFAULT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) houses() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE houses (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  houseable_id bigint(20) unsigned NOT NULL,
  houseable_type varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_houses_created_at (created_at),
  KEY idx_houses_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE houses (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  houseable_id int NOT NULL,
  houseable_type varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE houses (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  houseable_id int NOT NULL,
  houseable_type varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE houses (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  houseable_id bigint NOT NULL,
  houseable_type varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) phones() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE phones (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  phoneable_id bigint(20) unsigned NOT NULL,
  phoneable_type varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_phones_created_at (created_at),
  KEY idx_phones_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE phones (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  phoneable_id int NOT NULL,
  phoneable_type varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE phones (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  phoneable_id int NOT NULL,
  phoneable_type varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE phones (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  phoneable_id bigint NOT NULL,
  phoneable_type varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) roleUser() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE role_user (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  role_id bigint(20) unsigned NOT NULL,
  user_id bigint(20) unsigned NOT NULL,
  PRIMARY KEY (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE role_user (
  id SERIAL PRIMARY KEY NOT NULL,
  role_id int NOT NULL,
  user_id int NOT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE role_user (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  role_id int NOT NULL,
  user_id int NOT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE role_user (
  id bigint NOT NULL IDENTITY(1,1),
  role_id bigint NOT NULL,
  user_id bigint NOT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}

func (r *testTables) schema() string {
	switch r.driver {
	case contractsdatabase.DriverMysql:
		return `
CREATE TABLE goravel.schemas (
  id bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  name varchar(255) NOT NULL,
  created_at datetime(3) NOT NULL,
  updated_at datetime(3) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_schemas_created_at (created_at),
  KEY idx_schemas_updated_at (updated_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;
`
	case contractsdatabase.DriverPostgres:
		return `
CREATE TABLE goravel.schemas (
  id SERIAL PRIMARY KEY NOT NULL,
  name varchar(255) NOT NULL,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL
);
`
	case contractsdatabase.DriverSqlite:
		return `
CREATE TABLE goravel.schemas (
  id integer PRIMARY KEY AUTOINCREMENT NOT NULL,
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);
`
	case contractsdatabase.DriverSqlserver:
		return `
CREATE TABLE goravel.schemas (
  id bigint NOT NULL IDENTITY(1,1),
  name varchar(255) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL,
  PRIMARY KEY (id)
);
`
	default:
		return ""
	}
}
