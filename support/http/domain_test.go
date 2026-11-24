package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/errors"
)

func TestParseDomain(t *testing.T) {
	for _, domain := range []string{
		"",
		" ",
		".",
		"https://example.com",
		"example.com:8080",
		"example.com/admin",
		"*.example.com",
		"admin..example.com",
		".example.com",
		"-admin.example.com",
		"admin-.example.com",
		"ad min.example.com",
		"a{account}.example.com",
		"{account}-api.example.com",
		"{}.example.com",
		"{1st}.example.com",
		"{account}.{account}.example.com",
		"{account}.{Account}.example.com",
	} {
		t.Run(domain, func(t *testing.T) {
			_, err := ParseDomain(domain)
			assert.EqualError(t, err, errors.RouteInvalidDomain.Args(domain).Error())
		})
	}

	domain, err := ParseDomain(" Admin.Example.com. ")
	require.NoError(t, err)
	assert.Equal(t, "admin.example.com", domain.String())
	assert.False(t, domain.HasParams())
	assert.Empty(t, domain.Params())
	assert.Equal(t, "admin.example.com", domain.Key())

	domain, err = ParseDomain("{account}.{Region}.Example.com")
	require.NoError(t, err)
	assert.Equal(t, "{account}.{Region}.example.com", domain.String())
	assert.True(t, domain.HasParams())
	assert.Equal(t, []string{"account", "Region"}, domain.Params())
	assert.Equal(t, "{}.{}.example.com", domain.Key())

	other, err := ParseDomain("{tenant}.{zone}.example.com")
	require.NoError(t, err)
	assert.Equal(t, domain.Key(), other.Key())
}

func TestDomainMatch(t *testing.T) {
	static, err := ParseDomain("admin.example.com")
	require.NoError(t, err)

	for host, expected := range map[string]bool{
		"admin.example.com":      true,
		"Admin.Example.COM":      true,
		"admin.example.com:3000": true,
		"admin.example.com.":     true,
		"adminXexample.com":      false,
		"www.admin.example.com":  false,
		"admin.example.com.evil": false,
		"[::1]:3000":             false,
		"":                       false,
	} {
		params, ok := static.Match(host)
		assert.Equal(t, expected, ok, host)
		assert.Nil(t, params, host)
	}

	pattern, err := ParseDomain("{account}.{region}.example.com")
	require.NoError(t, err)

	params, ok := pattern.Match("Goravel.eu.example.com:443")
	assert.True(t, ok)
	assert.Equal(t, map[string]string{"account": "goravel", "region": "eu"}, params)

	for _, host := range []string{"eu.example.com", "a.b.c.example.com", "goravel.eu.example.org", "goravel.euXexample.com"} {
		_, ok := pattern.Match(host)
		assert.False(t, ok, host)
	}
}

func TestNormalizeHost(t *testing.T) {
	for host, expected := range map[string]string{
		"admin.example.com":      "admin.example.com",
		"Admin.Example.com:3000": "admin.example.com",
		"admin.example.com.":     "admin.example.com",
		"[::1]:3000":             "::1",
		"[::1]":                  "[::1]",
		"":                       "",
	} {
		assert.Equal(t, expected, NormalizeHost(host), host)
	}

	assert.Zero(t, testing.AllocsPerRun(10, func() {
		NormalizeHost("admin.example.com")
	}))
}
