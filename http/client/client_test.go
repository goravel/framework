package client

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/http/client"
	"github.com/goravel/framework/foundation/json"
)

type ClientTestSuite struct {
	suite.Suite
	json foundation.Json
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (s *ClientTestSuite) SetupTest() {
	s.json = json.New()
}

func (s *ClientTestSuite) TestNewClient_ConfigAppliedCorrectly() {
	cfg := &client.Config{
		BaseUrl:             "https://goravel.dev",
		Timeout:             5 * time.Second,
		MaxIdleConns:        123,
		MaxIdleConnsPerHost: 45,
		MaxConnsPerHost:     67,
		IdleConnTimeout:     89 * time.Second,
	}

	c := NewClient("test_client", cfg, s.json)

	s.Equal("test_client", c.Name())
	s.Same(cfg, c.Config(), "Config pointer should be stored directly")

	httpClient := c.HTTPClient()
	s.NotNil(httpClient)
	s.Equal(5*time.Second, httpClient.Timeout, "Timeout was not applied to http.Client")

	transport, ok := httpClient.Transport.(*http.Transport)
	s.True(ok, "Transport should be of type *http.Transport")

	s.Equal(123, transport.MaxIdleConns)
	s.Equal(45, transport.MaxIdleConnsPerHost)
	s.Equal(67, transport.MaxConnsPerHost)
	s.Equal(89*time.Second, transport.IdleConnTimeout)
}

func (s *ClientTestSuite) TestNewClient_NilConfig_UsesDefaults() {
	c := NewClient("default_client", nil, s.json)

	s.NotNil(c.Config())
	s.NotNil(c.HTTPClient())

	transport, ok := c.HTTPClient().Transport.(*http.Transport)
	s.True(ok)
	s.NotNil(transport)
}

func (s *ClientTestSuite) TestNewClientWithError_PropagatesLazyError() {
	expectedErr := errors.New("config missing")

	c := newClientWithError(expectedErr)
	s.NotNil(c)
	s.NotNil(c.Config(), "Config should be initialized to empty struct to prevent panics")

	req := c.NewRequest()
	s.NotNil(req)

	// We don't care about the URL here because it shouldn't even reach the network
	resp, err := req.Get("https://example.com")

	s.Nil(resp)
	s.Error(err)
	s.Equal(expectedErr, err, "The specific initialization error was not propagated to the request")
}

func (s *ClientTestSuite) TestDeepCopy_TransportIsolation() {
	// Ensure that two clients do not share the same Transport pointer.
	// If they did, changing one would break the other.
	cfg := &client.Config{MaxIdleConns: 10}

	client1 := NewClient("client1", cfg, s.json)
	client2 := NewClient("client2", cfg, s.json)

	t1 := client1.HTTPClient().Transport.(*http.Transport)
	t2 := client2.HTTPClient().Transport.(*http.Transport)

	// Pointers should be different
	s.NotSame(t1, t2, "Clients are sharing the same Transport instance, this is not thread-safe")

	// Modifying one shouldn't affect the other (though they started with same config values)
	t1.MaxIdleConns = 999
	s.Equal(10, t2.MaxIdleConns, "Modifying Client 1 transport affected Client 2")
}
