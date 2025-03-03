package client

import (
	"net/http"
	"sync"

	"github.com/goravel/framework/contracts/http/client"
)

var (
	httpClient     *http.Client
	httpClientOnce sync.Once
)

func getHttpClient(config *client.Config) *http.Client {
	httpClientOnce.Do(func() {
		httpClient = &http.Client{
			Timeout: config.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        config.MaxIdleConns,
				MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
				MaxConnsPerHost:     config.MaxConnsPerHost,
				IdleConnTimeout:     config.IdleConnTimeout,
			},
		}
	})
	return httpClient
}
