package filesystem

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/str"
	supporttime "github.com/goravel/framework/support/time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

/*
 * S3 OSS
 * Document: https://github.com/awsdocs/aws-doc-sdk-examples/blob/main/gov2/s3/common/main.go
 * More: https://aws.github.io/aws-sdk-go-v2/docs/sdk-utilities/s3/#putobjectinput-body-field-ioreadseeker-vs-ioreader
 */

type S3 struct {
	ctx      context.Context
	instance *s3.Client
	bucket   string
	disk     string
	url      string
}

func NewS3(ctx context.Context, disk string) (*S3, error) {
	accessKeyId := facades.Config.GetString(fmt.Sprintf("filesystems.disks.%s.key", disk))
	accessKeySecret := facades.Config.GetString(fmt.Sprintf("filesystems.disks.%s.secret", disk))
	region := facades.Config.GetString(fmt.Sprintf("filesystems.disks.%s.region", disk))
	bucket := facades.Config.GetString(fmt.Sprintf("filesystems.disks.%s.bucket", disk))
	url := facades.Config.GetString(fmt.Sprintf("filesystems.disks.%s.url", disk))

	client := s3.New(s3.Options{
		Region:      region,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
	})

	return &S3{
		ctx:      ctx,
		instance: client,
		bucket:   bucket,
		disk:     disk,
		url:      url,
	}, nil
}

func (r *S3) AllDirectories(path string) ([]string, error) {
	var directories []string
	validPath := validPath(path)
	listObjsResponse, err := r.instance.ListObjectsV2(r.ctx, &s3.ListObjectsV2Input{
		Bucket:    aws.String(r.bucket),
		Delimiter: aws.String("/"),
		Prefix:    aws.String(validPath),
	})
	if err != nil {
		return nil, err
	}

	for _, commonPrefix := range listObjsResponse.CommonPrefixes {
		prefix := *commonPrefix.Prefix
		directories = append(directories, strings.ReplaceAll(prefix, validPath, ""))

		subDirectories, err := r.AllDirectories(*commonPrefix.Prefix)
		if err != nil {
			return nil, err
		}
		for _, subDirectory := range subDirectories {
			directories = append(directories, strings.ReplaceAll(prefix+subDirectory, validPath, ""))
		}
	}

	return directories, nil
}

func (r *S3) AllFiles(path string) ([]string, error) {
	var files []string
	validPath := validPath(path)
	listObjsResponse, err := r.instance.ListObjectsV2(r.ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(r.bucket),
		Prefix: aws.String(validPath),
	})
	if err != nil {
		return nil, err
	}
	for _, object := range listObjsResponse.Contents {
		file := *object.Key
		if !strings.HasSuffix(file, "/") {
			files = append(files, strings.ReplaceAll(file, validPath, ""))
		}
	}

	return files, nil
}

func (r *S3) Copy(originFile, targetFile string) error {
	_, err := r.instance.CopyObject(r.ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(r.bucket),
		CopySource: aws.String(r.bucket + "/" + originFile),
		Key:        aws.String(targetFile),
	})

	return err
}

func (r *S3) Delete(files ...string) error {
	var objectIdentifiers []types.ObjectIdentifier
	for _, file := range files {
		objectIdentifiers = append(objectIdentifiers, types.ObjectIdentifier{
			Key: aws.String(file),
		})
	}

	_, err := r.instance.DeleteObjects(r.ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(r.bucket),
		Delete: &types.Delete{
			Objects: objectIdentifiers,
			Quiet:   true,
		},
	})

	return err
}

func (r *S3) DeleteDirectory(directory string) error {
	if !strings.HasSuffix(directory, "/") {
		directory += "/"
	}

	listObjectsV2Response, err := r.instance.ListObjectsV2(r.ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(r.bucket),
		Prefix: aws.String(directory),
	})
	if err != nil {
		return err
	}
	if len(listObjectsV2Response.Contents) == 0 {
		return nil
	}

	for {
		for _, item := range listObjectsV2Response.Contents {
			_, err = r.instance.DeleteObject(r.ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(r.bucket),
				Key:    item.Key,
			})
			if err != nil {
				return err
			}
		}

		if listObjectsV2Response.IsTruncated {
			listObjectsV2Response, err = r.instance.ListObjectsV2(r.ctx, &s3.ListObjectsV2Input{
				Bucket:            aws.String(r.bucket),
				ContinuationToken: listObjectsV2Response.ContinuationToken,
			})
			if err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
}

func (r *S3) Directories(path string) ([]string, error) {
	var directories []string
	validPath := validPath(path)
	listObjsResponse, err := r.instance.ListObjectsV2(r.ctx, &s3.ListObjectsV2Input{
		Bucket:    aws.String(r.bucket),
		Delimiter: aws.String("/"),
		Prefix:    aws.String(validPath),
	})
	if err != nil {
		return nil, err
	}
	for _, commonPrefix := range listObjsResponse.CommonPrefixes {
		directories = append(directories, strings.ReplaceAll(*commonPrefix.Prefix, validPath, ""))
	}

	return directories, nil
}

func (r *S3) Exists(file string) bool {
	_, err := r.instance.HeadObject(r.ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(file),
	})

	return err == nil
}

func (r *S3) Files(path string) ([]string, error) {
	var files []string
	validPath := validPath(path)
	listObjsResponse, err := r.instance.ListObjectsV2(r.ctx, &s3.ListObjectsV2Input{
		Bucket:    aws.String(r.bucket),
		Delimiter: aws.String("/"),
		Prefix:    aws.String(validPath),
	})
	if err != nil {
		return nil, err
	}
	for _, object := range listObjsResponse.Contents {
		files = append(files, strings.ReplaceAll(*object.Key, validPath, ""))
	}

	return files, nil
}

func (r *S3) Get(file string) (string, error) {
	resp, err := r.instance.GetObject(r.ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(file),
	})
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	return string(data), nil
}

func (r *S3) MakeDirectory(directory string) error {
	if !strings.HasSuffix(directory, "/") {
		directory += "/"
	}

	return r.Put(directory, "")
}

func (r *S3) Missing(file string) bool {
	return !r.Exists(file)
}

func (r *S3) Move(oldFile, newFile string) error {
	if err := r.Copy(oldFile, newFile); err != nil {
		return err
	}

	return r.Delete(oldFile)
}

func (r *S3) Path(file string) string {
	return file
}

func (r *S3) Put(file string, content string) error {
	_, err := r.instance.PutObject(r.ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(file),
		Body:   strings.NewReader(content),
	})

	return err
}

func (r *S3) PutFile(filePath string, source filesystem.File) (string, error) {
	return r.PutFileAs(filePath, source, str.Random(40))
}

func (r *S3) PutFileAs(filePath string, source filesystem.File, name string) (string, error) {
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

func (r *S3) Size(file string) (int64, error) {
	resp, err := r.instance.HeadObject(r.ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(file),
	})
	if err != nil {
		return 0, err
	}

	return resp.ContentLength, nil
}

func (r *S3) TemporaryUrl(file string, time time.Time) (string, error) {
	presignClient := s3.NewPresignClient(r.instance)
	presignParams := &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(file),
	}
	presignDuration := func(po *s3.PresignOptions) {
		po.Expires = time.Sub(supporttime.Now())
	}
	presignResult, err := presignClient.PresignGetObject(r.ctx, presignParams, presignDuration)
	if err != nil {
		return "", err
	}

	return presignResult.URL, nil
}

func (r *S3) WithContext(ctx context.Context) filesystem.Driver {
	driver, err := NewS3(ctx, r.disk)
	if err != nil {
		facades.Log.Errorf("init %s disk fail: %+v", r.disk, err)
	}

	return driver
}

func (r *S3) Url(file string) string {
	return strings.TrimSuffix(r.url, "/") + "/" + strings.TrimPrefix(file, "/")
}
