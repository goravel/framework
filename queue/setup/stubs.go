package main

import (
	"strings"
)

type Stubs struct{}

func (s Stubs) Config(module string) string {
	content := `package config

import (
	"DummyModule/app/facades"
)

func init() {
	config := facades.Config()
	config.Add("queue", map[string]any{
		// Default Queue Connection Name
		"default": "sync",

		// Queue Connections
		//
		// Here you may configure the connection information for each server that is used by your application.
		// Drivers: "sync", "database", "custom"
		"connections": map[string]any{
			"sync": map[string]any{
				"driver": "sync",
			},
			"database": map[string]any{
				"driver":     "database",
				"connection": "postgres",
				"queue":      "default",
				"concurrent": 1,
			},
		},

		// Failed Queue Jobs
		//
		// These options configure the behavior of failed queue job logging so you
		// can control how and where failed jobs are stored.
		"failed": map[string]any{
			"database": config.Env("DB_CONNECTION", "postgres"),
			"table":    "failed_jobs",
		},
	})
}
`

	return strings.ReplaceAll(content, "DummyModule", module)
}

func (s Stubs) JobMigration(pkg, module string) (fileName, structName, content string) {
	content = `package DummyPackage

import (
	"github.com/goravel/framework/contracts/database/schema"

	"DummyModule/app/facades"
)

type M20210101000001CreateJobsTable struct{}

// Signature The unique signature for the migration.
func (r *M20210101000001CreateJobsTable) Signature() string {
	return "20210101000002_create_jobs_table"
}

// Up Run the migrations.
func (r *M20210101000001CreateJobsTable) Up() error {
	if !facades.Schema().HasTable("jobs") {
		if err := facades.Schema().Create("jobs", func(table schema.Blueprint) {
			table.ID()
			table.String("queue")
			table.LongText("payload")
			table.UnsignedTinyInteger("attempts").Default(0)
			table.DateTimeTz("reserved_at").Nullable()
			table.DateTimeTz("available_at")
			table.DateTimeTz("created_at").UseCurrent()
			table.Index("queue")
		}); err != nil {
			return err
		}
	}

	if !facades.Schema().HasTable("failed_jobs") {
		if err := facades.Schema().Create("failed_jobs", func(table schema.Blueprint) {
			table.ID()
			table.String("uuid")
			table.Text("connection")
			table.Text("queue")
			table.LongText("payload")
			table.LongText("exception")
			table.DateTimeTz("failed_at").UseCurrent()
			table.Unique("uuid")
		}); err != nil {
			return err
		}
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20210101000001CreateJobsTable) Down() error {
	if err := facades.Schema().DropIfExists("jobs"); err != nil {
		return err
	}

	if err := facades.Schema().DropIfExists("failed_jobs"); err != nil {
		return err
	}

	return nil
}
`

	content = strings.ReplaceAll(content, "DummyPackage", pkg)
	content = strings.ReplaceAll(content, "DummyModule", module)

	return "20210101000001_create_jobs_table.go", "M20210101000001CreateJobsTable{}", content
}

func (s Stubs) QueueFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/queue"
)

func Queue() queue.Queue {
	return App().MakeQueue()
}
`
}
