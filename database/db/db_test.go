package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractsdb "github.com/goravel/framework/contracts/database/db"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksdriver "github.com/goravel/framework/mocks/database/driver"
	mockslogger "github.com/goravel/framework/mocks/database/logger"
	"github.com/goravel/framework/support/carbon"
)

func TestTxSelectPassesParameterizedSQL(t *testing.T) {
	ctx := context.Background()
	now := carbon.Now()
	carbon.SetTestNow(now)
	defer carbon.ClearTestNow()

	parameterizedSQL := "SELECT * FROM users WHERE name = ?"
	explainedSQL := `SELECT * FROM users WHERE name = "John"`

	t.Run("slice destination uses SelectContext with placeholder", func(t *testing.T) {
		mockBuilder := mocksdb.NewTxBuilder(t)
		mockLogger := mockslogger.NewLogger(t)
		tx := &Tx{ctx: ctx, logger: mockLogger, txBuilder: mockBuilder}

		var users []TestUser
		mockBuilder.EXPECT().Explain(parameterizedSQL, "John").Return(explainedSQL).Once()
		mockBuilder.EXPECT().SelectContext(ctx, &users, parameterizedSQL, "John").Return(nil).Once()
		mockLogger.EXPECT().Trace(ctx, now, explainedSQL, int64(0), nil).Once()

		assert.NoError(t, tx.Select(&users, parameterizedSQL, "John"))
	})

	t.Run("struct destination uses GetContext with placeholder", func(t *testing.T) {
		mockBuilder := mocksdb.NewTxBuilder(t)
		mockLogger := mockslogger.NewLogger(t)
		tx := &Tx{ctx: ctx, logger: mockLogger, txBuilder: mockBuilder}

		var user TestUser
		mockBuilder.EXPECT().Explain(parameterizedSQL, "John").Return(explainedSQL).Once()
		mockBuilder.EXPECT().GetContext(ctx, &user, parameterizedSQL, "John").Return(nil).Once()
		mockLogger.EXPECT().Trace(ctx, now, explainedSQL, int64(1), nil).Once()

		assert.NoError(t, tx.Select(&user, parameterizedSQL, "John"))
	})
}

func newMockDriver(t *testing.T) *mocksdriver.Driver {
	pool := contractsdatabase.Pool{Writers: []contractsdatabase.Config{{Driver: "postgres", Connection: "primary"}}}
	driver := mocksdriver.NewDriver(t)
	driver.EXPECT().Pool().Return(pool).Once()
	driver.EXPECT().Grammar().Return(nil).Once()
	return driver
}

func TestNewTx_BuildsInstrument(t *testing.T) {
	// NewTx always builds the instrument and wraps the builder; telemetry is
	// resolved lazily on the first query, so NewTx itself never touches it.
	mockTxBuilder := mocksdb.NewTxBuilder(t)
	tx := NewTx(context.Background(), newMockDriver(t), mockslogger.NewLogger(t), nil, mockTxBuilder, &[]TxLog{})

	assert.NotNil(t, tx.instrument)
	assert.NotEqual(t, contractsdb.TxBuilder(mockTxBuilder), tx.txBuilder)
}
