package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/errors"
)

type FakeTransportTestSuite struct {
	suite.Suite
	mockBase *MockRoundTripper
}

func TestFakeTransportTestSuite(t *testing.T) {
	suite.Run(t, new(FakeTransportTestSuite))
}

func (s *FakeTransportTestSuite) SetupTest() {
	s.mockBase = new(MockRoundTripper)
}

func (s *FakeTransportTestSuite) TestRoundTrip_MockHit() {
	mocks := map[string]any{
		"http://example.com": "mocked_response",
	}
	state := NewFakeState(nil, mocks)
	transport := NewFakeTransport(state, s.mockBase, nil)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	s.NoError(err)

	resp, err := transport.RoundTrip(req)
	s.NoError(err)
	s.NotNil(resp)

	body, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Equal("mocked_response", string(body))

	s.mockBase.AssertNotCalled(s.T(), "RoundTrip", mock.Anything)
}

func (s *FakeTransportTestSuite) TestRoundTrip_Passthrough() {
	state := NewFakeState(nil, nil)
	transport := NewFakeTransport(state, s.mockBase, nil)

	req, err := http.NewRequest("GET", "http://real-world.com", nil)
	s.NoError(err)

	realResp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("real_response")),
	}

	s.mockBase.On("RoundTrip", req).Return(realResp, nil)

	resp, err := transport.RoundTrip(req)
	s.NoError(err)

	body, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.Equal("real_response", string(body))

	s.mockBase.AssertExpectations(s.T())
}

func (s *FakeTransportTestSuite) TestRoundTrip_PreventStray() {
	state := NewFakeState(nil, nil)
	state.PreventStrayRequests()

	transport := NewFakeTransport(state, s.mockBase, nil)

	req, err := http.NewRequest("GET", "http://stray-url.com", nil)
	s.NoError(err)

	resp, err := transport.RoundTrip(req)
	s.Nil(resp)
	s.ErrorIs(err, errors.HttpClientStrayRequest)

	s.mockBase.AssertNotCalled(s.T(), "RoundTrip", mock.Anything)
}

func (s *FakeTransportTestSuite) TestRoundTrip_Hydration() {
	mocks := map[string]any{
		"*": func(_ client.Request) client.Response {
			return NewResponseFactory(nil).Status(200)
		},
	}
	state := NewFakeState(nil, mocks)
	transport := NewFakeTransport(state, s.mockBase, nil)

	payload := []byte(`{"foo":"bar"}`)
	req, err := http.NewRequest("POST", "http://api.com", bytes.NewBuffer(payload))
	s.NoError(err)

	ctx := context.WithValue(req.Context(), clientNameKey, "github_client")
	req = req.WithContext(ctx)

	resp, err := transport.RoundTrip(req)
	s.NoError(err)
	s.NotNil(resp)

	found := state.AssertSent(func(r client.Request) bool {
		return r.ClientName() == "github_client" && r.Body() == `{"foo":"bar"}`
	})

	s.True(found, "Transport failed to hydrate request context or body correctly")
}

type MockRoundTripper struct {
	mock.Mock
}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}
