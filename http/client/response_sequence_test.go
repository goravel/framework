package client

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/foundation/json"
)

type ResponseSequenceTestSuite struct {
	suite.Suite
	factory *ResponseFactory
}

func TestResponseSequenceTestSuite(t *testing.T) {
	suite.Run(t, new(ResponseSequenceTestSuite))
}

func (s *ResponseSequenceTestSuite) SetupTest() {
	s.factory = NewResponseFactory(json.New())
}

func (s *ResponseSequenceTestSuite) TestSequence_Flow() {
	s.Run("Iterates through mixed types", func() {
		sequence := NewResponseSequence(s.factory)

		sequence.PushStatus(201)
		sequence.PushString("Hello", 200)
		sequence.Push(s.factory.Json(map[string]int{"id": 1}, 200))

		// Call 1
		resp1 := sequence.GetNext()
		s.NotNil(resp1)
		s.Equal(201, resp1.Status())

		// Call 2
		resp2 := sequence.GetNext()
		s.NotNil(resp2)
		s.Equal(200, resp2.Status())
		body2, _ := resp2.Body()
		s.Equal("Hello", body2)

		// Call 3
		resp3 := sequence.GetNext()
		s.NotNil(resp3)
		body3, _ := resp3.Json()
		s.Equal(float64(1), body3["id"])
	})
}

func (s *ResponseSequenceTestSuite) TestSequence_WithCount() {
	s.Run("Repeats response N times based on count", func() {
		sequence := NewResponseSequence(s.factory)

		// Push 500 Error -> 3 times
		sequence.PushStatus(http.StatusInternalServerError, 3)
		// Push 200 OK -> 1 time
		sequence.PushStatus(http.StatusOK)

		// Call 1 (500)
		s.Equal(http.StatusInternalServerError, sequence.GetNext().Status())
		// Call 2 (500)
		s.Equal(http.StatusInternalServerError, sequence.GetNext().Status())
		// Call 3 (500)
		s.Equal(http.StatusInternalServerError, sequence.GetNext().Status())

		// Call 4 (200) - Should switch now
		s.Equal(http.StatusOK, sequence.GetNext().Status())
	})
}

func (s *ResponseSequenceTestSuite) TestSequence_WhenEmpty_Default() {
	s.Run("Strict Mode: Returns nil when exhausted", func() {
		sequence := NewResponseSequence(s.factory)
		sequence.PushStatus(http.StatusTeapot)

		s.Equal(http.StatusTeapot, sequence.GetNext().Status())

		// Call again (exhausted)
		// Expect nil -> This triggers "HttpClientHandlerReturnedNil" in FakeTransport
		s.Nil(sequence.GetNext())
	})
}

func (s *ResponseSequenceTestSuite) TestSequence_WhenEmpty_Custom() {
	s.Run("Returns specific fallback response when exhausted", func() {
		sequence := NewResponseSequence(s.factory)

		// Sequence: 200
		sequence.PushStatus(http.StatusOK)

		// Fallback: 404
		sequence.WhenEmpty(s.factory.Status(http.StatusNotFound))

		// Consume 200
		s.Equal(http.StatusOK, sequence.GetNext().Status())

		// Now empty -> Expect Custom Fallback (404)
		resp1 := sequence.GetNext()
		s.NotNil(resp1)
		s.Equal(http.StatusNotFound, resp1.Status())

		resp2 := sequence.GetNext()
		s.NotNil(resp2)
		s.Equal(http.StatusNotFound, resp2.Status())
	})
}

func (s *ResponseSequenceTestSuite) TestSequence_NoResponses() {
	s.Run("Strict Mode: Returns nil if initialized empty", func() {
		sequence := NewResponseSequence(s.factory)

		s.Nil(sequence.GetNext())
	})
}
