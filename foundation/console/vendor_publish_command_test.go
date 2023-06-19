package console

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
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

	// Create temporary source and target directories for testing
	sourceData := "test"
	sourceDir, err := os.MkdirTemp("", "source")
	s.Nil(err)
	defer func(path string) {
		err = os.RemoveAll(path)
		if err != nil {
			panic(err)
		}
	}(sourceDir)

	targetDir, err := os.MkdirTemp("", "target")
	s.Nil(err)
	defer func(path string) {
		err = os.RemoveAll(path)
		if err != nil {
			panic(err)
		}
	}(targetDir)

	// source and target are directory
	sourceFile := filepath.Join(sourceDir, "test.txt")
	s.Nil(os.WriteFile(sourceFile, []byte(sourceData), 0644))
	targetDir = filepath.Join(targetDir, "test")

	result, err := command.publish(sourceDir, targetDir, false, false)
	s.Nil(err)
	s.Equal(1, len(result))

	targetFile := filepath.Join(targetDir, "test.txt")
	content, err := os.ReadFile(targetFile)
	s.Nil(err)
	s.Equal(sourceData, string(content))

	// source is file and target is directory
	sourceFile = filepath.Join(sourceDir, "test1.txt")
	s.Nil(os.WriteFile(sourceFile, []byte(sourceData), 0644))

	result, err = command.publish(sourceFile, targetDir, false, false)
	s.Nil(err)
	s.Equal(1, len(result))

	targetFile = filepath.Join(targetDir, "test1.txt")
	content, err = os.ReadFile(targetFile)
	s.Nil(err)
	s.Equal("test", string(content))

	// source and target are file
	sourceFile = filepath.Join(sourceDir, "test2.txt")
	s.Nil(os.WriteFile(sourceFile, []byte(sourceData), 0644))
	targetFile = filepath.Join(targetDir, "test3.txt")

	result, err = command.publish(sourceFile, targetFile, false, false)
	s.Nil(err)
	s.Equal(1, len(result))

	content, err = os.ReadFile(targetFile)
	s.Nil(err)
	s.Equal("test", string(content))
}

func (s *VendorPublishCommandTestSuite) TestPublishFile() {
	command := &VendorPublishCommand{}

	sourceData := "This is a test file."
	sourceFile := "./test_source.txt"
	targetFile := "./test_target.txt"

	// Create a test source file
	err := os.WriteFile(sourceFile, []byte(sourceData), 0644)
	s.Nil(err)

	// Ensure publishFile creates target file when it doesn't exist and 'existing' flag is set
	created, err := command.publishFile(sourceFile, targetFile, true, false)
	s.Nil(err)
	s.False(created)

	// Ensure publishFile returns false when target file already exists and 'force' flag is not set
	created, err = command.publishFile(sourceFile, targetFile, false, false)
	s.Nil(err)
	s.True(created)
	content, err := os.ReadFile(targetFile)
	s.Nil(err)
	s.Equal(string(content), sourceData)

	created, err = command.publishFile(sourceFile, targetFile, false, false)
	s.Nil(err)
	s.False(created)

	// Ensure publishFile overwrites target file when 'force' flag is set
	newSourceData := "This is a new test file."
	err = os.WriteFile(sourceFile, []byte(newSourceData), 0644)
	s.Nil(err)

	created, err = command.publishFile(sourceFile, targetFile, false, true)
	s.Nil(err)
	s.True(created)
	content, err = os.ReadFile(targetFile)
	s.Nil(err)
	s.Equal(string(content), newSourceData)

	// Clean up test files
	s.Nil(os.Remove(sourceFile))
	s.Nil(os.Remove(targetFile))
}
