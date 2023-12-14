package http

import "time"

type Cookie struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	Expires  time.Time
	MaxAge   int
	Secure   bool
	HttpOnly bool
	Raw      string
	SameSite string
}
