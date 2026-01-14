package client

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation/json"
)

type FakeSequenceTestSuite struct {
	suite.Suite
	json    foundation.Json
	factory *FakeResponse
}

func TestFakeSequenceTestSuite(t *testing.T) {
	suite.Run(t, new(FakeSequenceTestSuite))
}

func (s *FakeSequenceTestSuite) SetupTest() {
	s.json = json.New()
	s.factory = NewFakeResponse(s.json)
}

func (s *FakeSequenceTestSuite) TestSequence_Flow() {
	s.Run("Iterates through mixed types", func() {
		sequence := NewFakeSequence(s.json)

		sequence.PushStatus(201)
		sequence.PushString("Hello", 200)
		sequence.Push(s.factory.Json(map[string]int{"id": 1}, 200))

		// Call 1
		resp1 := sequence.getNext()
		s.NotNil(resp1)
		s.Equal(201, resp1.Status())

		// Call 2
		resp2 := sequence.getNext()
		s.NotNil(resp2)
		s.Equal(200, resp2.Status())
		body2, _ := resp2.Body()
		s.Equal("Hello", body2)

		// Call 3
		resp3 := sequence.getNext()
		s.NotNil(resp3)
		body3, err := resp3.Json()
		s.NoError(err)
		s.Equal(float64(1), body3["id"])
	})
}

func (s *FakeSequenceTestSuite) TestSequence_WithCount() {
	s.Run("Repeats response N times based on count", func() {
		sequence := NewFakeSequence(s.json)

		// Push 500 Error -> 3 times
		sequence.PushStatus(http.StatusInternalServerError, 3)
		// Push 200 OK -> 1 time
		sequence.PushStatus(http.StatusOK)

		// Call 1 (500)
		s.Equal(http.StatusInternalServerError, sequence.getNext().Status())
		// Call 2 (500)
		s.Equal(http.StatusInternalServerError, sequence.getNext().Status())
		// Call 3 (500)
		s.Equal(http.StatusInternalServerError, sequence.getNext().Status())

		// Call 4 (200) - Should switch now
		s.Equal(http.StatusOK, sequence.getNext().Status())
	})
}

func (s *FakeSequenceTestSuite) TestSequence_WhenEmpty_Default() {
	s.Run("Strict Mode: Returns nil when exhausted", func() {
		sequence := NewFakeSequence(s.json)
		sequence.PushStatus(http.StatusTeapot)

		s.Equal(http.StatusTeapot, sequence.getNext().Status())

		// Call again (exhausted)
		// Expect nil -> This triggers "HttpClientHandlerReturnedNil" in FakeTransport
		s.Nil(sequence.getNext())
	})
}

func (s *FakeSequenceTestSuite) TestSequence_WhenEmpty_Custom() {
	s.Run("Returns specific fallback response when exhausted", func() {
		sequence := NewFakeSequence(s.json)

		// Sequence: 200
		sequence.PushStatus(http.StatusOK)

		// Fallback: 404
		sequence.WhenEmpty(s.factory.Status(http.StatusNotFound))

		// Consume 200
		s.Equal(http.StatusOK, sequence.getNext().Status())

		// Now empty -> Expect Custom Fallback (404)
		resp1 := sequence.getNext()
		s.NotNil(resp1)
		s.Equal(http.StatusNotFound, resp1.Status())

		resp2 := sequence.getNext()
		s.NotNil(resp2)
		s.Equal(http.StatusNotFound, resp2.Status())
	})
}

func (s *FakeSequenceTestSuite) TestSequence_NoResponses() {
	s.Run("Strict Mode: Returns nil if initialized empty", func() {
		sequence := NewFakeSequence(s.json)

		s.Nil(sequence.getNext())
	})
}
