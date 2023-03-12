package filesystem

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gookit/color"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/config"
	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/file"
	supporttime "github.com/goravel/framework/support/time"
	testingdocker "github.com/goravel/framework/testing/docker"
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
	minioPool, minioResource, err := initMinioDocker()
	assert.Nil(t, err)

	var driver filesystem.Driver

	disks := []TestDisk{
		{
			disk: "local",
			url:  "http://localhost/storage",
		},
		{
			disk: "oss",
			url:  facades.Config.GetString("filesystems.disks.oss.url"),
		},
		{
			disk: "cos",
			url:  facades.Config.GetString("filesystems.disks.cos.url"),
		},
		{
			disk: "s3",
			url:  facades.Config.GetString("filesystems.disks.s3.url"),
		},
		{
			disk: "minio",
			url:  facades.Config.GetString("filesystems.disks.minio.url"),
		},
		{
			disk: "custom",
			url:  "http://localhost/storage",
		},
	}

	tests := []struct {
		name  string
		setup func(disk TestDisk)
	}{
		{
			name: "Put",
			setup: func(disk TestDisk) {
				assert.Nil(t, driver.Put("Put/1.txt", "Goravel"))
				assert.True(t, driver.Exists("Put/1.txt"))
				assert.True(t, driver.Missing("Put/2.txt"))
			},
		},
		{
			name: "Get",
			setup: func(disk TestDisk) {
				assert.Nil(t, driver.Put("Get/1.txt", "Goravel"))
				assert.True(t, driver.Exists("Get/1.txt"))
				data, err := driver.Get("Get/1.txt")
				assert.Nil(t, err)
				assert.Equal(t, "Goravel", data)
				length, err := driver.Size("Get/1.txt")
				assert.Nil(t, err)
				assert.Equal(t, int64(7), length)
			},
		},
		{
			name: "PutFile_Text",
			setup: func(disk TestDisk) {
				fileInfo, err := NewFile("./test.txt")
				assert.Nil(t, err)
				path, err := driver.PutFile("PutFile", fileInfo)
				assert.Nil(t, err)
				assert.True(t, driver.Exists(path))
				data, err := driver.Get(path)
				assert.Nil(t, err)
				assert.Equal(t, "Goravel", data)
			},
		},
		{
			name: "PutFile_Image",
			setup: func(disk TestDisk) {
				fileInfo, err := NewFile("../logo.png")
				assert.Nil(t, err)
				path, err := driver.PutFile("PutFile", fileInfo)
				assert.Nil(t, err)
				assert.True(t, driver.Exists(path))
			},
		},
		{
			name: "PutFileAs_Text",
			setup: func(disk TestDisk) {
				fileInfo, err := NewFile("./test.txt")
				assert.Nil(t, err)
				path, err := driver.PutFileAs("PutFileAs", fileInfo, "text")
				assert.Nil(t, err)
				assert.Equal(t, "PutFileAs/text.txt", path)
				assert.True(t, driver.Exists(path))
				data, err := driver.Get(path)
				assert.Nil(t, err)
				assert.Equal(t, "Goravel", data)

				path, err = driver.PutFileAs("PutFileAs", fileInfo, "text1.txt")
				assert.Nil(t, err)
				assert.Equal(t, "PutFileAs/text1.txt", path)
				assert.True(t, driver.Exists(path))
				data, err = driver.Get(path)
				assert.Nil(t, err)
				assert.Equal(t, "Goravel", data)
			},
		},
		{
			name: "PutFileAs_Image",
			setup: func(disk TestDisk) {
				fileInfo, err := NewFile("../logo.png")
				assert.Nil(t, err)
				path, err := driver.PutFileAs("PutFileAs", fileInfo, "image")
				assert.Nil(t, err)
				assert.Equal(t, "PutFileAs/image.png", path)
				assert.True(t, driver.Exists(path))

				path, err = driver.PutFileAs("PutFileAs", fileInfo, "image1.png")
				assert.Nil(t, err)
				assert.Equal(t, "PutFileAs/image1.png", path)
				assert.True(t, driver.Exists(path))
			},
		},
		{
			name: "Url",
			setup: func(disk TestDisk) {
				assert.Nil(t, driver.Put("Url/1.txt", "Goravel"))
				assert.True(t, driver.Exists("Url/1.txt"))
				url := disk.url + "/Url/1.txt"
				assert.Equal(t, url, driver.Url("Url/1.txt"))
				if disk.disk != "local" && disk.disk != "custom" {
					resp, err := http.Get(url)
					assert.Nil(t, err)
					content, err := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					assert.Nil(t, err)
					assert.Equal(t, "Goravel", string(content))
				}
			},
		},
		{
			name: "TemporaryUrl",
			setup: func(disk TestDisk) {
				assert.Nil(t, driver.Put("TemporaryUrl/1.txt", "Goravel"))
				assert.True(t, driver.Exists("TemporaryUrl/1.txt"))
				url, err := driver.TemporaryUrl("TemporaryUrl/1.txt", supporttime.Now().Add(5*time.Second))
				assert.Nil(t, err)
				assert.NotEmpty(t, url)
				if disk.disk != "local" && disk.disk != "custom" {
					resp, err := http.Get(url)
					assert.Nil(t, err)
					content, err := ioutil.ReadAll(resp.Body)
					resp.Body.Close()
					assert.Nil(t, err)
					assert.Equal(t, "Goravel", string(content))
				}
			},
		},
		{
			name: "Copy",
			setup: func(disk TestDisk) {
				assert.Nil(t, driver.Put("Copy/1.txt", "Goravel"))
				assert.True(t, driver.Exists("Copy/1.txt"))
				assert.Nil(t, driver.Copy("Copy/1.txt", "Copy1/1.txt"))
				assert.True(t, driver.Exists("Copy/1.txt"))
				assert.True(t, driver.Exists("Copy1/1.txt"))
			},
		},
		{
			name: "Move",
			setup: func(disk TestDisk) {
				assert.Nil(t, driver.Put("Move/1.txt", "Goravel"))
				assert.True(t, driver.Exists("Move/1.txt"))
				assert.Nil(t, driver.Move("Move/1.txt", "Move1/1.txt"))
				assert.True(t, driver.Missing("Move/1.txt"))
				assert.True(t, driver.Exists("Move1/1.txt"))
			},
		},
		{
			name: "Delete",
			setup: func(disk TestDisk) {
				assert.Nil(t, driver.Put("Delete/1.txt", "Goravel"))
				assert.True(t, driver.Exists("Delete/1.txt"))
				assert.Nil(t, driver.Delete("Delete/1.txt"))
				assert.True(t, driver.Missing("Delete/1.txt"))
			},
		},
		{
			name: "MakeDirectory",
			setup: func(disk TestDisk) {
				assert.Nil(t, driver.MakeDirectory("MakeDirectory1/"))
				assert.Nil(t, driver.MakeDirectory("MakeDirectory2"))
				assert.Nil(t, driver.MakeDirectory("MakeDirectory3/MakeDirectory4"))
			},
		},
		{
			name: "DeleteDirectory",
			setup: func(disk TestDisk) {
				assert.Nil(t, driver.Put("DeleteDirectory/1.txt", "Goravel"))
				assert.True(t, driver.Exists("DeleteDirectory/1.txt"))
				assert.Nil(t, driver.DeleteDirectory("DeleteDirectory"))
				assert.True(t, driver.Missing("DeleteDirectory/1.txt"))
			},
		},
		{
			name: "Files",
			setup: func(disk TestDisk) {
				assert.Nil(t, driver.Put("Files/1.txt", "Goravel"))
				assert.Nil(t, driver.Put("Files/2.txt", "Goravel"))
				assert.Nil(t, driver.Put("Files/3/3.txt", "Goravel"))
				assert.Nil(t, driver.Put("Files/3/4/4.txt", "Goravel"))
				assert.True(t, driver.Exists("Files/1.txt"))
				assert.True(t, driver.Exists("Files/2.txt"))
				assert.True(t, driver.Exists("Files/3/3.txt"))
				assert.True(t, driver.Exists("Files/3/4/4.txt"))
				files, err := driver.Files("Files")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt"}, files)
				files, err = driver.Files("./Files")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt"}, files)
				files, err = driver.Files("/Files")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt"}, files)
				files, err = driver.Files("./Files/")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt"}, files)
			},
		},
		{
			name: "AllFiles",
			setup: func(disk TestDisk) {
				assert.Nil(t, driver.Put("AllFiles/1.txt", "Goravel"))
				assert.Nil(t, driver.Put("AllFiles/2.txt", "Goravel"))
				assert.Nil(t, driver.Put("AllFiles/3/3.txt", "Goravel"))
				assert.Nil(t, driver.Put("AllFiles/3/4/4.txt", "Goravel"))
				assert.True(t, driver.Exists("AllFiles/1.txt"))
				assert.True(t, driver.Exists("AllFiles/2.txt"))
				assert.True(t, driver.Exists("AllFiles/3/3.txt"))
				assert.True(t, driver.Exists("AllFiles/3/4/4.txt"))
				files, err := driver.AllFiles("AllFiles")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
				files, err = driver.AllFiles("./AllFiles")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
				files, err = driver.AllFiles("/AllFiles")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
				files, err = driver.AllFiles("./AllFiles/")
				assert.Nil(t, err)
				assert.Equal(t, []string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
			},
		},
		{
			name: "Directories",
			setup: func(disk TestDisk) {
				assert.Nil(t, driver.Put("Directories/1.txt", "Goravel"))
				assert.Nil(t, driver.Put("Directories/2.txt", "Goravel"))
				assert.Nil(t, driver.Put("Directories/3/3.txt", "Goravel"))
				assert.Nil(t, driver.Put("Directories/3/5/5.txt", "Goravel"))
				assert.Nil(t, driver.MakeDirectory("Directories/3/4"))
				assert.True(t, driver.Exists("Directories/1.txt"))
				assert.True(t, driver.Exists("Directories/2.txt"))
				assert.True(t, driver.Exists("Directories/3/3.txt"))
				assert.True(t, driver.Exists("Directories/3/4/"))
				assert.True(t, driver.Exists("Directories/3/5/5.txt"))
				files, err := driver.Directories("Directories")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/"}, files)
				files, err = driver.Directories("./Directories")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/"}, files)
				files, err = driver.Directories("/Directories")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/"}, files)
				files, err = driver.Directories("./Directories/")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/"}, files)
			},
		},
		{
			name: "AllDirectories",
			setup: func(disk TestDisk) {
				assert.Nil(t, driver.Put("AllDirectories/1.txt", "Goravel"))
				assert.Nil(t, driver.Put("AllDirectories/2.txt", "Goravel"))
				assert.Nil(t, driver.Put("AllDirectories/3/3.txt", "Goravel"))
				assert.Nil(t, driver.Put("AllDirectories/3/5/6/6.txt", "Goravel"))
				assert.Nil(t, driver.MakeDirectory("AllDirectories/3/4"))
				assert.True(t, driver.Exists("AllDirectories/1.txt"))
				assert.True(t, driver.Exists("AllDirectories/2.txt"))
				assert.True(t, driver.Exists("AllDirectories/3/3.txt"))
				assert.True(t, driver.Exists("AllDirectories/3/4/"))
				assert.True(t, driver.Exists("AllDirectories/3/5/6/6.txt"))
				files, err := driver.AllDirectories("AllDirectories")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
				files, err = driver.AllDirectories("./AllDirectories")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
				files, err = driver.AllDirectories("/AllDirectories")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
				files, err = driver.AllDirectories("./AllDirectories/")
				assert.Nil(t, err)
				assert.Equal(t, []string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
			},
		},
	}

	for _, disk := range disks {
		var err error
		driver, err = NewDriver(disk.disk)
		assert.NotNil(t, driver)
		assert.Nil(t, err)

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				test.setup(disk)
			})
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
			file.Remove("./storage")
		}
	}

	file.Remove("test.txt")
	assert.Nil(t, minioPool.Purge(minioResource))
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
			"minio": map[string]any{
				"driver":   "minio",
				"key":      application.Env("MINIO_ACCESS_KEY_ID"),
				"secret":   application.Env("MINIO_ACCESS_KEY_SECRET"),
				"region":   application.Env("MINIO_REGION"),
				"bucket":   application.Env("MINIO_BUCKET"),
				"url":      application.Env("MINIO_URL"),
				"endpoint": application.Env("MINIO_ENDPOINT"),
				"ssl":      application.Env("MINIO_SSL", false),
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

func initMinioDocker() (*dockertest.Pool, *dockertest.Resource, error) {
	pool, err := testingdocker.Pool()
	if err != nil {
		return nil, nil, err
	}

	key := facades.Config.GetString("filesystems.disks.minio.key")
	secret := facades.Config.GetString("filesystems.disks.minio.secret")
	bucket := facades.Config.GetString("filesystems.disks.minio.bucket")
	resource, err := testingdocker.Resource(pool, &dockertest.RunOptions{
		Repository: "minio/minio",
		Tag:        "latest",
		Env: []string{
			"MINIO_ACCESS_KEY=" + key,
			"MINIO_SECRET_KEY=" + secret,
		},
		Cmd: []string{
			"server",
			"/data",
		},
		ExposedPorts: []string{
			"9000/tcp",
		},
	})
	if err != nil {
		return nil, nil, err
	}

	_ = resource.Expire(600)

	endpoint := fmt.Sprintf("127.0.0.1:%s", resource.GetPort("9000/tcp"))
	facades.Config.Add("filesystems.disks.minio", map[string]any{
		"driver":   "minio",
		"key":      facades.Config.Env("MINIO_ACCESS_KEY_ID"),
		"secret":   facades.Config.Env("MINIO_ACCESS_KEY_SECRET"),
		"region":   facades.Config.Env("MINIO_REGION"),
		"bucket":   bucket,
		"url":      fmt.Sprintf("http://%s/%s", endpoint, bucket),
		"endpoint": endpoint,
		"ssl":      facades.Config.Env("MINIO_SSL", false),
	})

	if err := pool.Retry(func() error {
		client, err := minio.New(endpoint, &minio.Options{
			Creds: credentials.NewStaticV4(key, secret, ""),
		})
		if err != nil {
			return err
		}

		if err := client.MakeBucket(context.Background(), facades.Config.GetString("filesystems.disks.minio.bucket"), minio.MakeBucketOptions{}); err != nil {
			return err
		}

		policy := `{
        "Version": "2012-10-17",
        "Statement": [
            {
                "Action": [
                    "s3:GetObject",
                    "s3:PutObject"
                ],
                "Effect": "Allow",
                "Principal": "*",
                "Resource": [
                    "arn:aws:s3:::` + bucket + `/*"
                ]
            },
            {
                "Action": [
                    "s3:ListBucket"
                ],
                "Effect": "Allow",
                "Principal": "*",
                "Resource": [
                    "arn:aws:s3:::` + bucket + `"
                ]
            }
        ]
    }`

		if err := client.SetBucketPolicy(context.Background(), bucket, policy); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	return pool, resource, nil
}
