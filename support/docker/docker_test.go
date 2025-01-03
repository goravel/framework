package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"

	contractstesting "github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/support/env"
)

func TestDatabase(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skip test that using Docker")
	}

	tests := []struct {
		name          string
		containerType ContainerType
		num           int
		setup         func(drivers []contractstesting.DatabaseDriver)
	}{
		{
			name:          "num is 0",
			containerType: ContainerTypeMysql,
			num:           0,
		},
		{
			name:          "single postgres",
			containerType: ContainerTypePostgres,
			num:           1,
		},
		{
			name:          "multiple postgres",
			containerType: ContainerTypePostgres,
			num:           2,
		},
	}

	if TestModel == TestModelNormal {
		tests = append(tests, []struct {
			name          string
			containerType ContainerType
			num           int
			setup         func(drivers []contractstesting.DatabaseDriver)
		}{
			{
				name:          "single mysql",
				containerType: ContainerTypeMysql,
				num:           1,
			},
			{
				name:          "multiple mysql",
				containerType: ContainerTypeMysql,
				num:           2,
			},
			{
				name:          "single sqlite",
				containerType: ContainerTypeSqlite,
				num:           1,
			},
			{
				name:          "multiple sqlite",
				containerType: ContainerTypeSqlite,
				num:           2,
			},
			{
				name:          "single sqlserver",
				containerType: ContainerTypeSqlserver,
				num:           1,
			},
			{
				name:          "multiple sqlserver",
				containerType: ContainerTypeSqlserver,
				num:           2,
			},
		}...)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.num == 0 {
				assert.Panics(t, func() {
					Database(test.containerType, test.num)
				})
			} else {
				drivers := Database(test.containerType, test.num)

				assert.Len(t, drivers, test.num)

				if test.containerType == ContainerTypeSqlite {
					for _, driver := range drivers {
						assert.NoError(t, driver.Shutdown())
					}
				}
			}
		})
	}
}
