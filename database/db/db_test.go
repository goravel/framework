package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	contractsdatabase "github.com/goravel/framework/contracts/database"
	contractsdb "github.com/goravel/framework/contracts/database/db"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksdriver "github.com/goravel/framework/mocks/database/driver"
	mockslogger "github.com/goravel/framework/mocks/database/logger"
	mockstelemetry "github.com/goravel/framework/mocks/telemetry"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/telemetry"
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
	t.Run("nil when telemetry is off", func(t *testing.T) {
		tx := NewTx(context.Background(), newMockDriver(t), mockslogger.NewLogger(t), nil, nil, &[]TxLog{})
		assert.Nil(t, tx.instrument)
	})

	t.Run("built when telemetry is enabled", func(t *testing.T) {
		mockTelemetry := mockstelemetry.NewTelemetry(t)
		mockTelemetry.EXPECT().Tracer(mock.Anything).Return(tracenoop.NewTracerProvider().Tracer("test")).Maybe()
		mockTelemetry.EXPECT().Meter(mock.Anything).Return(metricnoop.NewMeterProvider().Meter("test")).Maybe()
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetBool("telemetry.instrumentation.database.enabled", true).Return(true).Maybe()

		originalFacade, originalConfig := telemetry.Facade, telemetry.ConfigFacade
		telemetry.Facade, telemetry.ConfigFacade = mockTelemetry, mockConfig
		defer func() { telemetry.Facade, telemetry.ConfigFacade = originalFacade, originalConfig }()

		tx := NewTx(context.Background(), newMockDriver(t), mockslogger.NewLogger(t), nil, nil, &[]TxLog{})
		assert.NotNil(t, tx.instrument)
	})

	t.Run("wraps tx builder when telemetry is enabled", func(t *testing.T) {
		mockTelemetry := mockstelemetry.NewTelemetry(t)
		mockTelemetry.EXPECT().Tracer(mock.Anything).Return(tracenoop.NewTracerProvider().Tracer("test")).Maybe()
		mockTelemetry.EXPECT().Meter(mock.Anything).Return(metricnoop.NewMeterProvider().Meter("test")).Maybe()
		mockConfig := mocksconfig.NewConfig(t)
		mockConfig.EXPECT().GetBool("telemetry.instrumentation.database.enabled", true).Return(true).Maybe()

		originalFacade, originalConfig := telemetry.Facade, telemetry.ConfigFacade
		telemetry.Facade, telemetry.ConfigFacade = mockTelemetry, mockConfig
		defer func() { telemetry.Facade, telemetry.ConfigFacade = originalFacade, originalConfig }()

		mockTxBuilder := mocksdb.NewTxBuilder(t)
		tx := NewTx(context.Background(), newMockDriver(t), mockslogger.NewLogger(t), nil, mockTxBuilder, &[]TxLog{})

		assert.NotNil(t, tx.instrument)
		assert.NotEqual(t, contractsdb.TxBuilder(mockTxBuilder), tx.txBuilder)
	})

	t.Run("passes tx builder through when telemetry is off", func(t *testing.T) {
		mockTxBuilder := mocksdb.NewTxBuilder(t)
		tx := NewTx(context.Background(), newMockDriver(t), mockslogger.NewLogger(t), nil, mockTxBuilder, &[]TxLog{})

		assert.Nil(t, tx.instrument)
		assert.Equal(t, contractsdb.TxBuilder(mockTxBuilder), tx.txBuilder)
	})
}
