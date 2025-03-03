package client

import "time"

type Config struct {
	BaseUrl             string
	Timeout             time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
	MaxConnsPerHost     int
}
