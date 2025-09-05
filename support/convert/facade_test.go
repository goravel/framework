package convert

import (
	"testing"

	"github.com/goravel/framework/contracts/binding"
	"github.com/stretchr/testify/assert"
)

func TestFacadeBindingConversion(t *testing.T) {
	assert.Equal(t, "Auth", BindingToFacade(binding.Auth))
	assert.Equal(t, binding.Auth, FacadeToBinding("Auth"))

	assert.Equal(t, "RateLimiter", BindingToFacade(binding.RateLimiter))
	assert.Equal(t, binding.RateLimiter, FacadeToBinding("RateLimiter"))

	assert.Equal(t, "DB", BindingToFacade(binding.DB))
	assert.Equal(t, binding.DB, FacadeToBinding("DB"))
}
