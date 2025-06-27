package console

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/console/command"
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

func (s *VendorPublishCommandTestSuite) TestSignature() {
	cmd := &VendorPublishCommand{}
	expected := "vendor:publish"
	s.Equal(expected, cmd.Signature(), "The signature should be 'vendor:publish'")
}

func (s *VendorPublishCommandTestSuite) TestDescription() {
	cmd := &VendorPublishCommand{}
	expected := "Publish any publishable assets from vendor packages"
	s.Require().Equal(expected, cmd.Description())
}

func (s *VendorPublishCommandTestSuite) TestExtend() {
	cmd := &VendorPublishCommand{}
	got := cmd.Extend()

	s.Run("should return correct category", func() {
		expected := "vendor"
		s.Require().Equal(expected, got.Category)
	})

	if len(got.Flags) > 0 {
		s.Run("should have correctly configured StringFlag", func() {
			flag, ok := got.Flags[0].(*command.BoolFlag)
			if !ok {
				s.Fail("existing flag is not BoolFlag (got type: %T)", got.Flags[0])
			}

			testCases := []struct {
				name     string
				got      any
				expected any
			}{
				{"Name", flag.Name, "existing"},
				{"Aliases", flag.Aliases, []string{"e"}},
				{"Usage", flag.Usage, "Publish and overwrite only the files that have already been published"},
			}

			for _, tc := range testCases {
				if !reflect.DeepEqual(tc.got, tc.expected) {
					s.Require().Equal(tc.expected, tc.got)
				}
			}
		})
	}
}

func (s *VendorPublishCommandTestSuite) TestGetSourceFiles() {
	cmd := &VendorPublishCommand{}

	sourceDir, err := os.MkdirTemp("", "source")
	s.Require().Nil(err)
	defer func(path string) {
		if err := file.Remove(path); err != nil {
			panic(err)
		}
	}(sourceDir)

	sourceFile := filepath.Join(sourceDir, "test.txt")
	s.Require().Nil(file.PutContent(sourceFile, "test"))
	sourceFile = filepath.Join(sourceDir, "dir1/test.txt")
	s.Require().Nil(file.PutContent(sourceFile, "test"))

	files, err := cmd.getSourceFiles(filepath.Join(sourceDir, "test.txt"))
	s.Require().NoError(err)
	s.ElementsMatch(files, []string{
		filepath.Join(sourceDir, "test.txt"),
	})

	files, err = cmd.getSourceFiles(sourceDir)
	s.Require().NoError(err)
	s.ElementsMatch(files, []string{
		filepath.Join(sourceDir, "test.txt"),
		filepath.Join(sourceDir, "dir1/test.txt"),
	})
}

func (s *VendorPublishCommandTestSuite) TestGetSourceFilesForDir() {
	cmd := &VendorPublishCommand{}

	sourceDir, err := os.MkdirTemp("", "source")
	s.Require().Nil(err)
	defer func(path string) {
		if err := file.Remove(path); err != nil {
			panic(err)
		}
	}(sourceDir)

	sourceFile := filepath.Join(sourceDir, "test.txt")
	s.Require().Nil(file.PutContent(sourceFile, "test"))
	sourceFile = filepath.Join(sourceDir, "test1.txt")
	s.Require().Nil(file.PutContent(sourceFile, "test"))
	sourceFile = filepath.Join(sourceDir, "dir1/test.txt")
	s.Require().Nil(file.PutContent(sourceFile, "test"))
	sourceFile = filepath.Join(sourceDir, "dir1/dir11/test.txt")
	s.Require().Nil(file.PutContent(sourceFile, "test"))
	sourceFile = filepath.Join(sourceDir, "dir2/test.txt")
	s.Require().Nil(file.PutContent(sourceFile, "test"))

	files, err := cmd.getSourceFiles(sourceDir)
	s.Require().NoError(err)
	s.ElementsMatch(files, []string{
		filepath.Join(sourceDir, "test.txt"),
		filepath.Join(sourceDir, "test1.txt"),
		filepath.Join(sourceDir, "dir1/test.txt"),
		filepath.Join(sourceDir, "dir1/dir11/test.txt"),
		filepath.Join(sourceDir, "dir2/test.txt"),
	})
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
			cmd := NewVendorPublishCommand(test.publishes, test.publishGroups)
			s.Equal(test.expectPaths, cmd.pathsForPackageOrGroup(test.packageName, test.group))
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
			cmd := NewVendorPublishCommand(test.publishes, test.publishGroups)
			s.Equal(test.expectPaths, cmd.pathsForProviderAndGroup(test.packageName, test.group))
		})
	}
}

func (s *VendorPublishCommandTestSuite) TestPublish() {
	cmd := &VendorPublishCommand{}

	// Create temporary source and target directories for testing
	sourceDir, err := os.MkdirTemp("", "source")
	s.Require().Nil(err)
	defer func(path string) {
		if err := file.Remove(path); err != nil {
			panic(err)
		}
	}(sourceDir)

	targetDir, err := os.MkdirTemp("", "target")
	s.Require().Nil(err)

	sourceFile := filepath.Join(sourceDir, "test.txt")
	s.Require().Nil(file.PutContent(sourceFile, "test"))
	sourceFile = filepath.Join(sourceDir, "test1.txt")
	s.Require().Nil(file.PutContent(sourceFile, "test"))
	sourceFile = filepath.Join(sourceDir, "dir1/test.txt")
	s.Require().Nil(file.PutContent(sourceFile, "test"))
	sourceFile = filepath.Join(sourceDir, "dir2/test.txt")
	s.Require().Nil(file.PutContent(sourceFile, "test"))

	// source and target are directory
	result, err := cmd.publish(sourceDir, targetDir, false, false)
	s.Require().Nil(err)
	s.Require().Equal(4, len(result))

	content, err := os.ReadFile(filepath.Join(targetDir, "test.txt"))
	s.Require().Nil(err)
	s.Equal("test", string(content))
	content, err = os.ReadFile(filepath.Join(targetDir, "test1.txt"))
	s.Require().Nil(err)
	s.Equal("test", string(content))
	content, err = os.ReadFile(filepath.Join(targetDir, "dir1/test.txt"))
	s.Require().Nil(err)
	s.Equal("test", string(content))
	content, err = os.ReadFile(filepath.Join(targetDir, "dir2/test.txt"))
	s.Require().Nil(err)
	s.Equal("test", string(content))

	s.Require().Nil(file.Remove(targetDir))

	// source is file and target is directory
	sourceFile = filepath.Join(sourceDir, "test.txt")
	result, err = cmd.publish(sourceFile, targetDir, false, false)
	s.Nil(err)
	s.Equal(1, len(result))

	content, err = os.ReadFile(filepath.Join(targetDir, "test.txt"))
	s.Require().Nil(err)
	s.Equal("test", string(content))

	s.Require().Nil(file.Remove(targetDir))

	// source and target are file
	sourceFile = filepath.Join(sourceDir, "test.txt")
	targetFile := filepath.Join(targetDir, "test.txt")

	result, err = cmd.publish(sourceFile, targetFile, false, false)
	s.Nil(err)
	s.Equal(1, len(result))

	content, err = os.ReadFile(targetFile)
	s.Require().Nil(err)
	s.Equal("test", string(content))

	s.Require().Nil(file.Remove(targetDir))
}

func (s *VendorPublishCommandTestSuite) TestPublishFile() {
	cmd := &VendorPublishCommand{}

	sourceData := "This is a test file."
	sourceFile := "./test_source.txt"
	targetFile := "./test_target.txt"

	// Create a test source file
	err := os.WriteFile(sourceFile, []byte(sourceData), 0644)
	s.Nil(err)

	// Ensure publishFile creates target file when it doesn't exist and 'existing' flag is set
	created, err := cmd.publishFile(sourceFile, targetFile, true, false)
	s.Nil(err)
	s.False(created)

	// Ensure publishFile returns false when target file already exists and 'force' flag is not set
	created, err = cmd.publishFile(sourceFile, targetFile, false, false)
	s.Nil(err)
	s.True(created)
	content, err := os.ReadFile(targetFile)
	s.Nil(err)
	s.Equal(string(content), sourceData)

	created, err = cmd.publishFile(sourceFile, targetFile, false, false)
	s.Nil(err)
	s.False(created)

	// Ensure publishFile overwrites target file when 'force' flag is set
	newSourceData := "This is a new test file."
	err = os.WriteFile(sourceFile, []byte(newSourceData), 0644)
	s.Nil(err)

	created, err = cmd.publishFile(sourceFile, targetFile, false, true)
	s.Nil(err)
	s.True(created)
	content, err = os.ReadFile(targetFile)
	s.Nil(err)
	s.Equal(string(content), newSourceData)

	// Clean up test files
	s.Nil(os.Remove(sourceFile))
	s.Nil(os.Remove(targetFile))
}
