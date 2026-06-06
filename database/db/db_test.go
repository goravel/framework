package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	mocksdb "github.com/goravel/framework/mocks/database/db"
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
