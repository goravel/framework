package modify

func commands() string {
	return `package bootstrap

import "github.com/goravel/framework/contracts/console"

func Commands() []console.Command {
	return []console.Command{}
}
`
}

func filters() string {
	return `package bootstrap

import "github.com/goravel/framework/contracts/validation"

func Filters() []validation.Filter {
	return []validation.Filter{}
}
`
}

func migrations() string {
	return `package bootstrap

import "github.com/goravel/framework/contracts/database/schema"

func Migrations() []schema.Migration {
	return []schema.Migration{}
}
`
}

func providers() string {
	return `package bootstrap

import "github.com/goravel/framework/contracts/foundation"

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{}
}
`
}

func jobs() string {
	return `package bootstrap

import "github.com/goravel/framework/contracts/queue"

func Jobs() []queue.Job {
	return []queue.Job{}
}
`
}

func seeders() string {
	return `package bootstrap

import "github.com/goravel/framework/contracts/database/seeder"

func Seeders() []seeder.Seeder {
	return []seeder.Seeder{}
}
`
}

func rules() string {
	return `package bootstrap

import "github.com/goravel/framework/contracts/validation"

func Rules() []validation.Rule {
	return []validation.Rule{}
}
`
}
