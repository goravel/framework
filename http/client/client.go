package client

import (
	"net/http"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/config"
)

var (
	httpClient     *http.Client
	httpClientOnce sync.Once
)

func getHttpClient(config config.Config) *http.Client {
	httpClientOnce.Do(func() {
		httpClient = &http.Client{
			Timeout: config.GetDuration("http.client.timeout", 30*time.Second),
			Transport: &http.Transport{
				MaxIdleConns:        config.GetInt("http.client.max_idle_conns"),
				MaxIdleConnsPerHost: config.GetInt("http.client.max_idle_conns_per_host"),
				MaxConnsPerHost:     config.GetInt("http.client.max_conns_per_host"),
				IdleConnTimeout:     config.GetDuration("http.client.idle_conn_timeout"),
			},
		}
	})
	return httpClient
}
