package access

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/auth/access"
)

type contextKey int

const key contextKey = 0

type GateTestSuite struct {
	suite.Suite
}

func TestGateTestSuite(t *testing.T) {
	suite.Run(t, new(GateTestSuite))
}

func (s *GateTestSuite) SetupTest() {

}

func (s *GateTestSuite) TestWithContext() {
	ctx := context.WithValue(context.Background(), key, "goravel")

	gate := NewGate(ctx)
	gate.Define("create", func(ctx context.Context, arguments map[string]any) access.Response {
		user := arguments["user"].(string)
		if user == "1" {
			return NewAllowResponse()
		} else {
			return NewDenyResponse(ctx.Value(key).(string))
		}
	})

	assert.Equal(s.T(), NewDenyResponse("goravel"), gate.Inspect("create", map[string]any{
		"user": "2",
	}))
}

func (s *GateTestSuite) TestAllows() {
	gate := initGate()
	assert.True(s.T(), gate.Allows("create", map[string]any{
		"user": "1",
	}))
	assert.False(s.T(), gate.Allows("create", map[string]any{
		"user": "2",
	}))
	assert.False(s.T(), gate.Allows("update", map[string]any{
		"user": "1",
	}))
}

func (s *GateTestSuite) TestDenies() {
	gate := initGate()
	assert.False(s.T(), gate.Denies("create", map[string]any{
		"user": "1",
	}))
	assert.True(s.T(), gate.Denies("create", map[string]any{
		"user": "2",
	}))
	assert.True(s.T(), gate.Denies("update", map[string]any{
		"user": "1",
	}))
}

func (s *GateTestSuite) TestInspect() {
	gate := initGate()
	assert.Equal(s.T(), NewAllowResponse(), gate.Inspect("create", map[string]any{
		"user": "1",
	}))
	assert.True(s.T(), gate.Inspect("create", map[string]any{
		"user": "1",
	}).Allowed())
	assert.Equal(s.T(), NewDenyResponse("create error"), gate.Inspect("create", map[string]any{
		"user": "2",
	}))
	assert.Equal(s.T(), "create error", gate.Inspect("create", map[string]any{
		"user": "2",
	}).Message())
	assert.Equal(s.T(), NewDenyResponse(fmt.Sprintf("ability doesn't exist: %s", "delete")), gate.Inspect("delete", map[string]any{
		"user": "1",
	}))
}

func (s *GateTestSuite) TestAny() {
	gate := initGate()
	assert.True(s.T(), gate.Any([]string{"create", "update"}, map[string]any{
		"user": "1",
	}))
	assert.True(s.T(), gate.Any([]string{"create", "update"}, map[string]any{
		"user": "2",
	}))
	assert.False(s.T(), gate.Any([]string{"create", "update"}, map[string]any{
		"user": "3",
	}))
}

func (s *GateTestSuite) TestNone() {
	gate := initGate()
	assert.False(s.T(), gate.None([]string{"create", "update"}, map[string]any{
		"user": "1",
	}))
	assert.False(s.T(), gate.None([]string{"create", "update"}, map[string]any{
		"user": "2",
	}))
	assert.True(s.T(), gate.None([]string{"create", "update"}, map[string]any{
		"user": "3",
	}))
}

func (s *GateTestSuite) TestBefore() {
	gate := initGate()
	gate.Before(func(ctx context.Context, ability string, arguments map[string]any) access.Response {
		user := arguments["user"].(string)
		if user == "3" {
			return NewAllowResponse()
		}

		return nil
	})
	assert.True(s.T(), gate.Allows("create", map[string]any{
		"user": "3",
	}))
	assert.False(s.T(), gate.Allows("create", map[string]any{
		"user": "4",
	}))
}

func (s *GateTestSuite) TestAfter() {
	gate := initGate()
	gate.Define("delete", func(ctx context.Context, arguments map[string]any) access.Response {
		user := arguments["user"].(string)
		if user == "3" {
			return nil
		} else {
			return NewAllowResponse()
		}
	})
	gate.After(func(ctx context.Context, ability string, arguments map[string]any, result access.Response) access.Response {
		user := arguments["user"].(string)
		if user == "3" {
			return NewAllowResponse()
		}

		return nil
	})
	assert.True(s.T(), gate.Allows("delete", map[string]any{
		"user": "1",
	}))
	assert.True(s.T(), gate.Allows("delete", map[string]any{
		"user": "3",
	}))
}

func initGate() *Gate {
	gate := NewGate(context.Background())
	gate.Define("create", func(ctx context.Context, arguments map[string]any) access.Response {
		user := arguments["user"].(string)
		if user == "1" {
			return NewAllowResponse()
		} else {
			return NewDenyResponse("create error")
		}
	})
	gate.Define("update", func(ctx context.Context, arguments map[string]any) access.Response {
		user := arguments["user"].(string)
		if user == "2" {
			return NewAllowResponse()
		} else {
			return NewDenyResponse(" update error")
		}
	})

	return gate
}
