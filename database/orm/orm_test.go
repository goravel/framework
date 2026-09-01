package orm_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/contracts/database"
	contractsorm "github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/database/orm"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
)

func TestOrmConnectionResolvesPerConnectionConfig(t *testing.T) {
	o := newTestOrm(t)

	got := o.Connection("sqlite")
	require.Equal(t, "sqlite", got.Name())
	require.Equal(t, "sqlite_db", got.DatabaseName())
	require.Equal(t, "sqlite", got.Config().Connection)
}

func TestOrmConnectionWarmPathConcurrentReads(t *testing.T) {
	o := newTestOrm(t)

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)

	var mu sync.Mutex
	names := make([]string, 0, n)
	databaseNames := make([]string, 0, n)

	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			got := o.Connection("sqlite")
			mu.Lock()
			names = append(names, got.Name())
			databaseNames = append(databaseNames, got.DatabaseName())
			mu.Unlock()
		}()
	}

	wg.Wait()

	require.Len(t, names, n)
	require.Len(t, databaseNames, n)
	for _, name := range names {
		require.Equal(t, "sqlite", name)
	}
	for _, databaseName := range databaseNames {
		require.Equal(t, "sqlite_db", databaseName)
	}
}

func newTestOrm(t *testing.T) *orm.Orm {
	t.Helper()

	defaultCfg := database.Config{Connection: "postgres", Database: "default_db"}
	sqliteCfg := database.Config{Connection: "sqlite", Database: "sqlite_db"}
	defaultQuery := mocksorm.NewQuery(t)
	sqliteQuery := mocksorm.NewQuery(t)

	cache := orm.NewQueriesCache(
		map[string]contractsorm.Query{"postgres": defaultQuery, "sqlite": sqliteQuery},
		map[string]database.Config{"postgres": defaultCfg, "sqlite": sqliteCfg},
	)

	return orm.NewOrm(context.Background(), nil, "postgres", defaultCfg, defaultQuery, cache, nil, nil, nil, nil)
}
