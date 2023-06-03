package console

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/support/file"
)

type VendorPublishCommandTestSuite struct {
	suite.Suite
}

func TestVendorPublishCommandTestSuite(t *testing.T) {
	suite.Run(t, new(VendorPublishCommandTestSuite))
}

func (s *VendorPublishCommandTestSuite) SetupTest() {

}

func (s *VendorPublishCommandTestSuite) TestPathsForPackageOrGroup() {
	tests := []struct {
		name          string
		packageName   string
		group         string
		publishes     map[string]map[string]string
		publishGroups map[string]map[string]string
		expectPaths   map[string]string
	}{
		{
			name: "packageName and group are empty",
		},
		{
			name:        "packageName and group are not empty, and have same path",
			packageName: "github.com/goravel/sms",
			group:       "public",
			publishes: map[string]map[string]string{
				"github.com/goravel/sms": {
					"config.go": "config.go",
				},
			},
			publishGroups: map[string]map[string]string{
				"public": {
					"config.go": "config.go",
				},
			},
			expectPaths: map[string]string{
				"config.go": "config.go",
			},
		},
		{
			name:        "packageName and group are not empty, but doesn't have same path",
			packageName: "github.com/goravel/sms",
			group:       "public",
			publishes: map[string]map[string]string{
				"github.com/goravel/sms": {
					"config.go": "config.go",
				},
			},
			publishGroups: map[string]map[string]string{
				"public": {
					"config1.go": "config.go",
				},
			},
			expectPaths: map[string]string{},
		},
		{
			name:  "packageName is empty, group is not empty",
			group: "public",
			publishes: map[string]map[string]string{
				"github.com/goravel/sms": {
					"config.go": "config.go",
				},
			},
			publishGroups: map[string]map[string]string{
				"public": {
					"config1.go": "config.go",
				},
			},
			expectPaths: map[string]string{
				"config1.go": "config.go",
			},
		},
		{
			name:        "packageName is not empty, group is empty",
			packageName: "github.com/goravel/sms",
			publishes: map[string]map[string]string{
				"github.com/goravel/sms": {
					"config.go": "config.go",
				},
			},
			publishGroups: map[string]map[string]string{
				"public": {
					"config1.go": "config.go",
				},
			},
			expectPaths: map[string]string{
				"config.go": "config.go",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			command := NewVendorPublishCommand(test.publishes, test.publishGroups)
			s.Equal(test.expectPaths, command.pathsForPackageOrGroup(test.packageName, test.group))
		})
	}
}

func (s *VendorPublishCommandTestSuite) TestPathsForProviderAndGroup() {
	tests := []struct {
		name          string
		packageName   string
		group         string
		publishes     map[string]map[string]string
		publishGroups map[string]map[string]string
		expectPaths   map[string]string
	}{
		{
			name:        "not found packageName",
			packageName: "github.com/goravel/sms1",
			group:       "public",
			publishes: map[string]map[string]string{
				"github.com/goravel/sms": {
					"config.go": "config.go",
				},
			},
			publishGroups: map[string]map[string]string{
				"public": {
					"config1.go": "config.go",
				},
			},
		},
		{
			name:        "not found group",
			packageName: "github.com/goravel/sms",
			group:       "public1",
			publishes: map[string]map[string]string{
				"github.com/goravel/sms": {
					"config.go": "config.go",
				},
			},
			publishGroups: map[string]map[string]string{
				"public": {
					"config1.go": "config.go",
				},
			},
		},
		{
			name:        "does not have Intersection",
			packageName: "github.com/goravel/sms",
			group:       "public",
			publishes: map[string]map[string]string{
				"github.com/goravel/sms": {
					"config.go": "config.go",
				},
			},
			publishGroups: map[string]map[string]string{
				"public": {
					"config1.go": "config.go",
				},
			},
			expectPaths: map[string]string{},
		},
		{
			name:        "have Intersection",
			packageName: "github.com/goravel/sms",
			group:       "public",
			publishes: map[string]map[string]string{
				"github.com/goravel/sms": {
					"config.go": "config.go",
				},
			},
			publishGroups: map[string]map[string]string{
				"public": {
					"config.go": "config.go",
				},
			},
			expectPaths: map[string]string{
				"config.go": "config.go",
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			command := NewVendorPublishCommand(test.publishes, test.publishGroups)
			s.Equal(test.expectPaths, command.pathsForProviderAndGroup(test.packageName, test.group))
		})
	}
}

func (s *VendorPublishCommandTestSuite) TestPublish() {
	command := &VendorPublishCommand{}

	success, err := command.publish("test.go", "123", true, false)
	s.False(success)
	s.Nil(err)
	success, err = command.publish("test.go", "123", false, false)
	s.True(success)
	s.Nil(err)
	s.True(file.Contain("test.go", "123"))
	success, err = command.publish("test.go", "123", false, false)
	s.False(success)
	s.Nil(err)
	success, err = command.publish("test.go", "111", false, true)
	s.True(success)
	s.Nil(err)
	s.True(file.Contain("test.go", "111"))
	success, err = command.publish("test.go", "222", true, false)
	s.True(success)
	s.Nil(err)
	s.True(file.Contain("test.go", "222"))
	success, err = command.publish("test.go", "333", true, true)
	s.True(success)
	s.Nil(err)
	s.True(file.Contain("test.go", "333"))
	s.True(file.Remove("test.go"))
}
