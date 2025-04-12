package driver

import (
	"os"
	"path/filepath"
	"testing"

	sessioncontract "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
	"github.com/stretchr/testify/suite"
)

type FileTestSuite struct {
	suite.Suite
}

func TestFileTestSuite(t *testing.T) {
	suite.Run(t, &FileTestSuite{})
}

func (f *FileTestSuite) BeforeTest(suiteName, testName string) {

	err := file.Remove(f.getPath())

	f.Require().True(err == nil || os.IsNotExist(err), "Failed to clean test directory '%s' before test: %v", f.getPath(), err)
}

func (f *FileTestSuite) AfterTest(suiteName, testName string) {
	err := file.Remove(f.getPath())
	f.Require().True(err == nil || os.IsNotExist(err), "Failed to clean test directory '%s' after test: %v", f.getPath(), err)
}

func (f *FileTestSuite) TestNewFile_EmptyPathError() {
	driver, err := newFile("", f.getMinutes())
	f.Error(err, "newFile should return error for empty path")
	f.Nil(driver, "Driver should be nil on error")
	f.Contains(err.Error(), "session file path cannot be empty", "Error message mismatch")
}

func (f *FileTestSuite) TestClose() {
	driver := f.getDriver()
	f.Require().NotNil(driver)
	f.Nil(driver.Close())
}

func (f *FileTestSuite) TestDestroy() {
	driver := f.getDriver()

	f.Nil(driver.Destroy("foo"))

	f.Nil(driver.Write("foo", "bar"))
	value, err := driver.Read("foo")
	f.Nil(err)
	f.Equal("bar", value)

	f.Nil(driver.Destroy("foo"))

	value, err = driver.Read("foo")
	f.Nil(err)
	f.Equal("", value)
}

func (f *FileTestSuite) TestGc() {
	driver := f.getDriver()
	f.Require().NotNil(driver)
	lifetimeSeconds := f.getMinutes() * 60

	sessionIDValid := "gc_valid_session"
	f.Nil(driver.Write(sessionIDValid, "this session should survive gc"))

	f.Nil(driver.Gc(lifetimeSeconds))

	valueValid, errValid := driver.Read(sessionIDValid)
	f.Nil(errValid)
	f.Equal("this session should survive gc", valueValid, "Valid session removed by GC")

	sessionIDExpired := "gc_expired_session"
	f.Nil(driver.Write(sessionIDExpired, "this session should be removed by gc"))
	f.True(file.Exists(filepath.Join(f.getPath(), sessionIDValid)), "Expired session file missing after GC")

	carbon.SetTestNow(carbon.Now(carbon.UTC).AddMinutes(f.getMinutes()).AddMinutes(20))
	defer carbon.UnsetTestNow()

	f.Nil(driver.Gc(lifetimeSeconds))

	valueExpired, errExpired := driver.Read(sessionIDExpired)
	f.Nil(errExpired, "Read on GC'd session should not error")
	f.Equal("", valueExpired, "Expired session not removed by GC")
	f.False(file.Exists(filepath.Join(f.getPath(), sessionIDExpired)), "Valid session file still exists after GC")
}

func (f *FileTestSuite) TestGc_NonExistentPath() {
	driver := f.getDriver()
	f.Require().NotNil(driver)
	lifetimeSeconds := f.getMinutes() * 60

	err := file.Remove(f.getPath())
	f.Require().True(err == nil || os.IsNotExist(err), "Failed to remove test directory during test setup")
	f.False(file.Exists(f.getPath()), "Test directory should be gone before calling Gc")

	gcErr := driver.Gc(lifetimeSeconds)
	f.Nil(gcErr, "Gc should not return error for non-existent base path")
}

func (f *FileTestSuite) TestOpen() {
	driver := f.getDriver()
	f.Require().NotNil(driver)
	f.Nil(driver.Open("", ""))
}

func (f *FileTestSuite) TestRead() {
	driver := f.getDriver()
	f.Require().NotNil(driver)
	sessionID := "read_test_session"
	sessionData := "data to be read"

	value, err := driver.Read("read_non_existent")
	f.Nil(err)
	f.Equal("", value, "Reading non-existent session should return empty string")

	f.Nil(driver.Write(sessionID, sessionData))
	value, err = driver.Read(sessionID)
	f.Nil(err)
	f.Equal(sessionData, value, "Failed to read back recently written data")

	carbon.SetTestNow(carbon.Now(carbon.UTC).AddMinutes(f.getMinutes()).AddSeconds(-1))
	value, err = driver.Read(sessionID)
	f.Nil(err)
	f.Equal(sessionData, value, "Session expired too early")
	carbon.UnsetTestNow()

	carbon.SetTestNow(carbon.Now(carbon.UTC).AddMinutes(f.getMinutes()).AddSeconds(1))
	value, err = driver.Read(sessionID)
	f.Nil(err, "Read on expired session should not error")
	f.Equal("", value, "Session did not expire correctly")
	carbon.UnsetTestNow()
}

func (f *FileTestSuite) TestWrite() {
	driver := f.getDriver()
	f.Require().NotNil(driver)
	sessionID := "write_test_session"

	f.Nil(driver.Write(sessionID, "initial data"))
	value, err := driver.Read(sessionID)
	f.Nil(err)
	f.Equal("initial data", value)

	f.True(file.Exists(filepath.Join(f.getPath(), sessionID)))

	f.Nil(driver.Write(sessionID, "overwritten data"))
	value, err = driver.Read(sessionID)
	f.Nil(err)
	f.Equal("overwritten data", value)
}

func BenchmarkFile_ReadWrite(b *testing.B) {

	f := new(FileTestSuite)
	f.SetT(&testing.T{})
	f.BeforeTest("", "")

	driver := f.getDriver()
	require := f.Require()
	require.NotNil(driver)

	sessionID := "bench_session_id"
	sessionData := "benchmark data"

	err := driver.Write(sessionID, sessionData)
	require.Nil(err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		errWrite := driver.Write(sessionID, sessionData)
		if errWrite != nil {
			b.Fatalf("Write failed during benchmark: %v", errWrite)
		}

		value, errRead := driver.Read(sessionID)
		if errRead != nil {
			b.Fatalf("Read failed during benchmark: %v", errRead)
		}
		if value != sessionData {
			b.Fatalf("Read returned incorrect data during benchmark: got %s, want %s", value, sessionData)
		}
	}
	b.StopTimer()

}

func (f *FileTestSuite) getDriver() sessioncontract.Driver {
	driver, err := newFile(f.getPath(), f.getMinutes())

	f.Require().NoError(err, "Failed to create file driver for test")
	f.Require().NotNil(driver, "Created driver is nil")
	return driver
}

func (f *FileTestSuite) getPath() string {

	return "storage/framework/sessions_test"
}

func (f *FileTestSuite) getMinutes() int {
	return 5
}
