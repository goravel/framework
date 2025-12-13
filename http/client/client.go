package client

import (
	"net/http"
	"sync"

	"github.com/goravel/framework/contracts/http/client"
)

var (
	httpTransport  *http.Transport
	httpClientOnce sync.Once
)

func getHttpClient(config *client.Config) *http.Client {
	httpClientOnce.Do(func() {
		httpTransport = &http.Transport{
			MaxIdleConns:        config.MaxIdleConns,
			MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
			MaxConnsPerHost:     config.MaxConnsPerHost,
			IdleConnTimeout:     config.IdleConnTimeout,
		}
	})

	return &http.Client{
		Timeout:   config.Timeout,
		Transport: httpTransport,
	}
}
