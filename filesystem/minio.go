package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	neturl "net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/str"
	supporttime "github.com/goravel/framework/support/time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

/*
 * MinIO OSS
 * Document: https://min.io/docs/minio/linux/developers/minio-drivers.html#go-sdk
 * More: https://min.io/docs/minio/linux/developers/go/minio-go.html
 * More: https://min.io/docs/minio/linux/developers/go/API.html
 * Example: https://github.com/minio/minio-go/tree/master/examples
 * Example: https://github.com/minio/minio-go/tree/master/examples/s3
 */

// Minio v1.0.0
type Minio struct {
	ctx      context.Context
	instance *minio.Client
	bucket   string
	disk     string
	url      string
}

// NewMinio v1.0.0
func NewMinio(ctx context.Context, disk string) (*Minio, error) {
	key := facades.Config.GetString(fmt.Sprintf("filesystems.disks.%s.key", disk))
	secret := facades.Config.GetString(fmt.Sprintf("filesystems.disks.%s.secret", disk))
	region := facades.Config.GetString(fmt.Sprintf("filesystems.disks.%s.region", disk))
	bucket := facades.Config.GetString(fmt.Sprintf("filesystems.disks.%s.bucket", disk))
	url := facades.Config.GetString(fmt.Sprintf("filesystems.disks.%s.url", disk))
	endpoint := facades.Config.GetString(fmt.Sprintf("filesystems.disks.%s.endpoint", disk))
	useSSL := facades.Config.GetBool(fmt.Sprintf("filesystems.disks.%s.use_ssl", disk), false)
	autoCreateBucket := facades.Config.GetBool(fmt.Sprintf("filesystems.disks.%s.auto_create_bucket", disk), false)

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(key, secret, ""),
		Secure: useSSL,
		Region: region, // Distributed use
	})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("[filesystem] init %s driver error: %+v", disk, err))
	}

	// Auto create bucket
	if autoCreateBucket {
		exists, errBucketExists := client.BucketExists(ctx, bucket)
		if errBucketExists == nil && exists {
		} else {
			err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: region})
			if err != nil {
				return nil, errors.New(fmt.Sprintf("[filesystem] %s driver auto create bucket error: %+v", disk, err))
			}
		}
	}

	return &Minio{
		ctx:      ctx,
		instance: client,
		bucket:   bucket,
		disk:     disk,
		url:      url,
	}, nil
}

// AllDirectories v1.0.0
func (r *Minio) AllDirectories(path string) ([]string, error) {
	var directories []string
	validPath := validPath(path)
	objectCh := r.instance.ListObjects(r.ctx, r.bucket, minio.ListObjectsOptions{
		Prefix:    validPath,
		Recursive: false, // Whether the recursive | ignore '/' delimiter
	})

	wg := sync.WaitGroup{}
	for object := range objectCh {
		if object.Err != nil {
			continue
		}
		prefix := object.Key
		directories = append(directories, strings.ReplaceAll(prefix, validPath, ""))

		wg.Add(1)
		subDirectories, err := r.AllDirectories(prefix)
		if err != nil {
			return nil, err
		}
		for _, subDirectory := range subDirectories {
			if strings.HasSuffix(subDirectory, "/") {
				directories = append(directories, strings.ReplaceAll(prefix+subDirectory, validPath, ""))
			}
		}
		wg.Done()
	}
	wg.Wait()

	return directories, nil
}

// AllFiles v1.0.0
func (r *Minio) AllFiles(path string) ([]string, error) {
	var files []string
	validPath := validPath(path)
	objectsCh := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectsCh)
		for object := range r.instance.ListObjects(r.ctx, r.bucket, minio.ListObjectsOptions{
			Prefix:    validPath,
			Recursive: true,
		}) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}
			objectsCh <- object
		}
	}()
	for object := range objectsCh {
		if object.Err != nil {
			return nil, object.Err
		}
		filename := object.Key
		if !strings.HasSuffix(filename, "/") {
			files = append(files, strings.ReplaceAll(filename, validPath, ""))
		}
	}
	return files, nil
}

// Copy v1.0.0
func (r *Minio) Copy(originFile, targetFile string) error {
	srcOpts := minio.CopySrcOptions{
		Bucket: r.bucket,
		Object: originFile,
	}
	dstOpts := minio.CopyDestOptions{
		Bucket: r.bucket,
		Object: targetFile,
	}
	_, err := r.instance.CopyObject(r.ctx, dstOpts, srcOpts)
	return err
}

// Delete v1.0.0
func (r *Minio) Delete(files ...string) error {
	objectsCh := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectsCh)
		for _, filename := range files {
			object := minio.ObjectInfo{
				Key: filename,
			}
			objectsCh <- object
		}
	}()
	opts := minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	}
	for rErr := range r.instance.RemoveObjects(r.ctx, r.bucket, objectsCh, opts) {
		return rErr.Err
	}

	return nil
}

// DeleteDirectory v1.0.0
func (r *Minio) DeleteDirectory(directory string) error {
	if !strings.HasSuffix(directory, "/") {
		directory += "/"
	}
	opts := minio.RemoveObjectOptions{
		ForceDelete:      true,
		GovernanceBypass: true,
		VersionID:        "",
	}
	err := r.instance.RemoveObject(r.ctx, r.bucket, directory, opts)
	if err != nil {
		return err
	}

	return nil
}

// Directories v1.0.0
func (r *Minio) Directories(path string) ([]string, error) {
	var directories []string
	validPath := validPath(path)
	objectCh := r.instance.ListObjects(r.ctx, r.bucket, minio.ListObjectsOptions{
		Prefix:    validPath,
		Recursive: false, // Whether the recursive | ignore '/' delimiter
	})
	for object := range objectCh {
		if object.Err != nil {
			continue
		}
		prefix := object.Key
		if strings.HasSuffix(prefix, "/") {
			directories = append(directories, strings.ReplaceAll(prefix, validPath, ""))
		}
	}

	return directories, nil
}

// Exists v1.0.0
func (r *Minio) Exists(file string) bool {
	_, err := r.instance.StatObject(r.ctx, r.bucket, file, minio.StatObjectOptions{})
	if err != nil {
		return false
	}

	return true
}

// Files v1.0.0
func (r *Minio) Files(path string) ([]string, error) {
	var files []string
	validPath := validPath(path)
	objectsCh := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectsCh)
		for object := range r.instance.ListObjects(r.ctx, r.bucket, minio.ListObjectsOptions{
			Prefix:    validPath,
			Recursive: false,
		}) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}
			objectsCh <- object
		}
	}()
	for object := range objectsCh {
		if object.Err != nil {
			return nil, object.Err
		}
		filename := object.Key
		if !strings.HasSuffix(filename, "/") {
			files = append(files, strings.ReplaceAll(filename, validPath, ""))
		}
	}
	return files, nil
}

// Get v1.0.0
func (r *Minio) Get(file string) (string, error) {
	object, err := r.instance.GetObject(r.ctx, r.bucket, file, minio.GetObjectOptions{})
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(object)
	defer object.Close()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MakeDirectory v1.0.0
func (r *Minio) MakeDirectory(directory string) error {
	if !strings.HasSuffix(directory, "/") {
		directory += "/"
	}

	return r.Put(directory, "")
}

// Missing v1.0.0
func (r *Minio) Missing(file string) bool {
	return !r.Exists(file)
}

// Move v1.0.0
func (r *Minio) Move(oldFile, newFile string) error {
	if err := r.Copy(oldFile, newFile); err != nil {
		return err
	}

	return r.Delete(oldFile)
}

// Path v1.0.0
func (r *Minio) Path(file string) string {
	return file
}

// Put v1.0.0
func (r *Minio) Put(file string, content string) error {
	_, err := r.instance.PutObject(
		r.ctx, r.bucket,
		file, strings.NewReader(content),
		strings.NewReader(content).Size(),
		minio.PutObjectOptions{},
	)
	return err
}

// PutFile v1.0.0
func (r *Minio) PutFile(filePath string, source filesystem.File) (string, error) {
	return r.PutFileAs(filePath, source, str.Random(40))
}

// PutFileAs v1.0.0
func (r *Minio) PutFileAs(filePath string, source filesystem.File, name string) (string, error) {
	fullPath, err := fullPathOfFile(filePath, source, name)
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadFile(source.File())
	if err != nil {
		return "", err
	}

	if err := r.Put(fullPath, string(data)); err != nil {
		return "", err
	}

	return fullPath, nil
}

// Size v1.0.0
func (r *Minio) Size(file string) (int64, error) {
	objInfo, err := r.instance.StatObject(r.ctx, r.bucket, file, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}
	return objInfo.Size, nil
}

// TemporaryUrl v1.0.0
func (r *Minio) TemporaryUrl(file string, time time.Time) (string, error) {
	file = strings.TrimPrefix(file, "/")
	reqParams := make(neturl.Values)
	fileBaseName := filepath.Base(file)
	reqParams.Set("response-content-disposition", "attachment; filename=\""+fileBaseName+"\"")
	presignedURL, err := r.instance.PresignedGetObject(r.ctx, r.bucket, file, time.Sub(supporttime.Now()), reqParams)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(r.url, "/") + presignedURL.Path, nil
}

// WithContext v1.0.0
func (r *Minio) WithContext(ctx context.Context) filesystem.Driver {
	driver, err := NewMinio(ctx, r.disk)
	if err != nil {
		facades.Log.Errorf("init %s disk fail: %+v", r.disk, err)
	}

	return driver
}

// Url v1.0.0
func (r *Minio) Url(file string) string {
	return strings.TrimSuffix(r.url, "/") + "/" + r.bucket + "/" + strings.TrimPrefix(file, "/")
}
