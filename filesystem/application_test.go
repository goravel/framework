package filesystem

import (
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/file"
	supporttime "github.com/goravel/framework/support/time"
)

type TestDisk struct {
	disk string
	url  string
}

func TestStorage(t *testing.T) {
	if !file.Exists("../.env") {
		color.Redln("No filesystem tests run, need create .env based on .env.example, then initialize it")
		return
	}

	file.Create("test.txt", "Goravel")
	initConfig()

	var driver filesystem.Driver

	disks := []TestDisk{
		{
			disk: "local",
			url:  "http://localhost/storage/",
		},
		{
			disk: "oss",
			url:  "https://goravel.oss-cn-beijing.aliyuncs.com/",
		},
		{
			disk: "cos",
			url:  "https://goravel-1257814968.cos.ap-beijing.myqcloud.com/",
		},
		{
			disk: "s3",
			url:  "https://goravel.s3.us-east-2.amazonaws.com/",
		},
		{
			disk: "custom",
			url:  "http://localhost/storage/",
		},
	}

	tests := []struct {
		name  string
		setup func(name string, disk TestDisk)
	}{
		{
			name: "Put",
			setup: func(name string, disk TestDisk) {
				assert.Nil(t, driver.Put("Put/1.txt", "Goravel"), name)
				assert.True(t, driver.Exists("Put/1.txt"), name)
				assert.True(t, driver.Missing("Put/2.txt"), name)
			},
		},
		{
			name: "Get",
			setup: func(name string, disk TestDisk) {
				assert.Nil(t, driver.Put("Get/1.txt", "Goravel"), name)
				assert.True(t, driver.Exists("Get/1.txt"), name)
				data, err := driver.Get("Get/1.txt")
				assert.Nil(t, err, name)
				assert.Equal(t, "Goravel", data, name)
				length, err := driver.Size("Get/1.txt")
				assert.Nil(t, err, name)
				assert.Equal(t, int64(7), length, name)
			},
		},
		{
			name: "PutFile_Text",
			setup: func(name string, disk TestDisk) {
				fileInfo, err := NewFile("./test.txt")
				assert.Nil(t, err, name)
				path, err := driver.PutFile("PutFile", fileInfo)
				assert.Nil(t, err, name)
				assert.True(t, driver.Exists(path), name)
				data, err := driver.Get(path)
				assert.Nil(t, err, name)
				assert.Equal(t, "Goravel", data, name)
			},
		},
		{
			name: "PutFile_Image",
			setup: func(name string, disk TestDisk) {
				fileInfo, err := NewFile("../logo.png")
				assert.Nil(t, err, name)
				path, err := driver.PutFile("PutFile", fileInfo)
				assert.Nil(t, err, name)
				assert.True(t, driver.Exists(path), name)
			},
		},
		{
			name: "PutFileAs_Text",
			setup: func(name string, disk TestDisk) {
				fileInfo, err := NewFile("./test.txt")
				assert.Nil(t, err, name)
				path, err := driver.PutFileAs("PutFileAs", fileInfo, "text")
				assert.Nil(t, err, name)
				assert.Equal(t, "PutFileAs/text.txt", path, name)
				assert.True(t, driver.Exists(path), name)
				data, err := driver.Get(path)
				assert.Nil(t, err, name)
				assert.Equal(t, "Goravel", data, name)

				path, err = driver.PutFileAs("PutFileAs", fileInfo, "text1.txt")
				assert.Nil(t, err, name)
				assert.Equal(t, "PutFileAs/text1.txt", path, name)
				assert.True(t, driver.Exists(path), name)
				data, err = driver.Get(path)
				assert.Nil(t, err, name)
				assert.Equal(t, "Goravel", data, name)
			},
		},
		{
			name: "PutFileAs_Image",
			setup: func(name string, disk TestDisk) {
				fileInfo, err := NewFile("../logo.png")
				assert.Nil(t, err, name)
				path, err := driver.PutFileAs("PutFileAs", fileInfo, "image")
				assert.Nil(t, err, name)
				assert.Equal(t, "PutFileAs/image.png", path, name)
				assert.True(t, driver.Exists(path), name)

				path, err = driver.PutFileAs("PutFileAs", fileInfo, "image1.png")
				assert.Nil(t, err, name)
				assert.Equal(t, "PutFileAs/image1.png", path, name)
				assert.True(t, driver.Exists(path), name)
			},
		},
		{
			name: "Url",
			setup: func(name string, disk TestDisk) {
				assert.Nil(t, driver.Put("Url/1.txt", "Goravel"), name)
				assert.True(t, driver.Exists("Url/1.txt"), name)
				assert.Equal(t, disk.url+"Url/1.txt", driver.Url("Url/1.txt"), name)
				if disk.disk != "local" && disk.disk != "custom" {
					resp, err := http.Get(disk.url + "Url/1.txt")
					assert.Nil(t, err, name)
					content, err := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					assert.Nil(t, err, name)
					assert.Equal(t, "Goravel", string(content), name)
				}
			},
		},
		{
			name: "TemporaryUrl",
			setup: func(name string, disk TestDisk) {
				assert.Nil(t, driver.Put("TemporaryUrl/1.txt", "Goravel"), name)
				assert.True(t, driver.Exists("TemporaryUrl/1.txt"), name)
				url, err := driver.TemporaryUrl("TemporaryUrl/1.txt", supporttime.Now().Add(5*time.Second))
				assert.Nil(t, err, name)
				assert.NotEmpty(t, url, name)
				if disk.disk != "local" && disk.disk != "custom" {
					resp, err := http.Get(url)
					assert.Nil(t, err, name)
					content, err := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					assert.Nil(t, err, name)
					assert.Equal(t, "Goravel", string(content), name)
				}
			},
		},
		{
			name: "Copy",
			setup: func(name string, disk TestDisk) {
				assert.Nil(t, driver.Put("Copy/1.txt", "Goravel"), name)
				assert.True(t, driver.Exists("Copy/1.txt"), name)
				assert.Nil(t, driver.Copy("Copy/1.txt", "Copy1/1.txt"), name)
				assert.True(t, driver.Exists("Copy/1.txt"), name)
				assert.True(t, driver.Exists("Copy1/1.txt"), name)
			},
		},
		{
			name: "Move",
			setup: func(name string, disk TestDisk) {
				assert.Nil(t, driver.Put("Move/1.txt", "Goravel"), name)
				assert.True(t, driver.Exists("Move/1.txt"), name)
				assert.Nil(t, driver.Move("Move/1.txt", "Move1/1.txt"), name)
				assert.True(t, driver.Missing("Move/1.txt"), name)
				assert.True(t, driver.Exists("Move1/1.txt"), name)
			},
		},
		{
			name: "Delete",
			setup: func(name string, disk TestDisk) {
				assert.Nil(t, driver.Put("Delete/1.txt", "Goravel"), name)
				assert.True(t, driver.Exists("Delete/1.txt"), name)
				assert.Nil(t, driver.Delete("Delete/1.txt"), name)
				assert.True(t, driver.Missing("Delete/1.txt"), name)
			},
		},
		{
			name: "MakeDirectory",
			setup: func(name string, disk TestDisk) {
				assert.Nil(t, driver.MakeDirectory("MakeDirectory1/"), name)
				assert.Nil(t, driver.MakeDirectory("MakeDirectory2"), name)
				assert.Nil(t, driver.MakeDirectory("MakeDirectory3/MakeDirectory4"), name)
			},
		},
		{
			name: "DeleteDirectory",
			setup: func(name string, disk TestDisk) {
				assert.Nil(t, driver.Put("DeleteDirectory/1.txt", "Goravel"), name)
				assert.True(t, driver.Exists("DeleteDirectory/1.txt"), name)
				assert.Nil(t, driver.DeleteDirectory("DeleteDirectory"), name)
				assert.True(t, driver.Missing("DeleteDirectory/1.txt"), name)
			},
		},
		{
			name: "Files",
			setup: func(name string, disk TestDisk) {
				assert.Nil(t, driver.Put("Files/1.txt", "Goravel"), name)
				assert.Nil(t, driver.Put("Files/2.txt", "Goravel"), name)
				assert.Nil(t, driver.Put("Files/3/3.txt", "Goravel"), name)
				assert.Nil(t, driver.Put("Files/3/4/4.txt", "Goravel"), name)
				assert.True(t, driver.Exists("Files/1.txt"), name)
				assert.True(t, driver.Exists("Files/2.txt"), name)
				assert.True(t, driver.Exists("Files/3/3.txt"), name)
				assert.True(t, driver.Exists("Files/3/4/4.txt"), name)
				files, err := driver.Files("Files")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"1.txt", "2.txt"}, files, name)
				files, err = driver.Files("./Files")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"1.txt", "2.txt"}, files, name)
				files, err = driver.Files("/Files")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"1.txt", "2.txt"}, files, name)
				files, err = driver.Files("./Files/")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"1.txt", "2.txt"}, files, name)
			},
		},
		{
			name: "AllFiles",
			setup: func(name string, disk TestDisk) {
				assert.Nil(t, driver.Put("AllFiles/1.txt", "Goravel"), name)
				assert.Nil(t, driver.Put("AllFiles/2.txt", "Goravel"), name)
				assert.Nil(t, driver.Put("AllFiles/3/3.txt", "Goravel"), name)
				assert.Nil(t, driver.Put("AllFiles/3/4/4.txt", "Goravel"), name)
				assert.True(t, driver.Exists("AllFiles/1.txt"), name)
				assert.True(t, driver.Exists("AllFiles/2.txt"), name)
				assert.True(t, driver.Exists("AllFiles/3/3.txt"), name)
				assert.True(t, driver.Exists("AllFiles/3/4/4.txt"), name)
				files, err := driver.AllFiles("AllFiles")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files, name)
				files, err = driver.AllFiles("./AllFiles")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files, name)
				files, err = driver.AllFiles("/AllFiles")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files, name)
				files, err = driver.AllFiles("./AllFiles/")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files, name)
			},
		},
		{
			name: "Directories",
			setup: func(name string, disk TestDisk) {
				assert.Nil(t, driver.Put("Directories/1.txt", "Goravel"), name)
				assert.Nil(t, driver.Put("Directories/2.txt", "Goravel"), name)
				assert.Nil(t, driver.Put("Directories/3/3.txt", "Goravel"), name)
				assert.Nil(t, driver.Put("Directories/3/5/5.txt", "Goravel"), name)
				assert.Nil(t, driver.MakeDirectory("Directories/3/4"), name)
				assert.True(t, driver.Exists("Directories/1.txt"), name)
				assert.True(t, driver.Exists("Directories/2.txt"), name)
				assert.True(t, driver.Exists("Directories/3/3.txt"), name)
				assert.True(t, driver.Exists("Directories/3/4/"), name)
				assert.True(t, driver.Exists("Directories/3/5/5.txt"), name)
				files, err := driver.Directories("Directories")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"3/"}, files, name)
				files, err = driver.Directories("./Directories")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"3/"}, files, name)
				files, err = driver.Directories("/Directories")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"3/"}, files, name)
				files, err = driver.Directories("./Directories/")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"3/"}, files, name)
			},
		},
		{
			name: "AllDirectories",
			setup: func(name string, disk TestDisk) {
				assert.Nil(t, driver.Put("AllDirectories/1.txt", "Goravel"), name)
				assert.Nil(t, driver.Put("AllDirectories/2.txt", "Goravel"), name)
				assert.Nil(t, driver.Put("AllDirectories/3/3.txt", "Goravel"), name)
				assert.Nil(t, driver.Put("AllDirectories/3/5/6/6.txt", "Goravel"), name)
				assert.Nil(t, driver.MakeDirectory("AllDirectories/3/4"), name)
				assert.True(t, driver.Exists("AllDirectories/1.txt"), name)
				assert.True(t, driver.Exists("AllDirectories/2.txt"), name)
				assert.True(t, driver.Exists("AllDirectories/3/3.txt"), name)
				assert.True(t, driver.Exists("AllDirectories/3/4/"), name)
				assert.True(t, driver.Exists("AllDirectories/3/5/6/6.txt"), name)
				files, err := driver.AllDirectories("AllDirectories")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"3/", "3/4/", "3/5/", "3/5/6/"}, files, name)
				files, err = driver.AllDirectories("./AllDirectories")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"3/", "3/4/", "3/5/", "3/5/6/"}, files, name)
				files, err = driver.AllDirectories("/AllDirectories")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"3/", "3/4/", "3/5/", "3/5/6/"}, files, name)
				files, err = driver.AllDirectories("./AllDirectories/")
				assert.Nil(t, err, name)
				assert.Equal(t, []string{"3/", "3/4/", "3/5/", "3/5/6/"}, files, name)
			},
		},
	}

	for _, disk := range disks {
		var err error
		driver, err = NewDriver(disk.disk)
		assert.NotNil(t, driver)
		assert.Nil(t, err)

		for _, test := range tests {
			test.setup(disk.disk+" "+test.name, disk)
		}

		assert.Nil(t, driver.DeleteDirectory("Put"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("Get"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("PutFile"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("PutFileAs"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("Url"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("TemporaryUrl"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("Copy"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("Copy1"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("Move"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("Move1"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("Delete"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("MakeDirectory1"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("MakeDirectory2"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("MakeDirectory3"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("MakeDirectory4"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("DeleteDirectory"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("Files"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("AllFiles"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("Directories"), disk.disk)
		assert.Nil(t, driver.DeleteDirectory("AllDirectories"), disk.disk)

		if disk.disk == "local" || disk.disk == "custom" {
			assert.True(t, file.Remove("./storage"))
		}
	}
	file.Remove("test.txt")
}

func initConfig() {
	application := config.NewApplication("../.env")
	application.Add("filesystems", map[string]any{
		"default": "local",
		"disks": map[string]any{
			"local": map[string]any{
				"driver": "local",
				"root":   "storage/app",
				"url":    "http://localhost/storage",
			},
			"s3": map[string]any{
				"driver": "s3",
				"key":    application.Env("AWS_ACCESS_KEY_ID"),
				"secret": application.Env("AWS_ACCESS_KEY_SECRET"),
				"region": application.Env("AWS_DEFAULT_REGION"),
				"bucket": application.Env("AWS_BUCKET"),
				"url":    application.Env("AWS_URL"),
			},
			"oss": map[string]any{
				"driver":   "oss",
				"key":      application.Env("ALIYUN_ACCESS_KEY_ID"),
				"secret":   application.Env("ALIYUN_ACCESS_KEY_SECRET"),
				"bucket":   application.Env("ALIYUN_BUCKET"),
				"url":      application.Env("ALIYUN_URL"),
				"endpoint": application.Env("ALIYUN_ENDPOINT"),
			},
			"cos": map[string]any{
				"driver": "cos",
				"key":    application.Env("TENCENT_ACCESS_KEY_ID"),
				"secret": application.Env("TENCENT_ACCESS_KEY_SECRET"),
				"bucket": application.Env("TENCENT_BUCKET"),
				"url":    application.Env("TENCENT_URL"),
			},
			"custom": map[string]any{
				"driver": "custom",
				"via": &Local{
					root: "storage/app/public",
					url:  "http://localhost/storage",
				},
			},
		},
	})

	facades.Config = application
}
