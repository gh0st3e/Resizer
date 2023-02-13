package localstack

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
)

const (
	Small  = "small"
	Medium = "medium"
	Large  = "large"

	path = "images/"
)

type LocalStack struct {
	logger  *logrus.Logger
	Session *session.Session
	Client  *s3.S3
	Bucket  string
}

func GetConnection(logger *logrus.Logger, bucket string) (*LocalStack, error) {
	var localStack LocalStack
	localStack.Bucket = bucket
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("us-east-1"),
		Credentials:      credentials.NewStaticCredentials(bucket, bucket, ""),
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String("http://localstack:4566"),
	})

	if err != nil {
		return nil, err
	}

	client := s3.New(sess)

	localStack.Session = sess
	localStack.Client = client
	localStack.logger = logger

	return &localStack, nil
}

func (l *LocalStack) SetDirName(msg string) string {
	return strings.SplitAfterN(msg, "-", 2)[0]

}

func (l *LocalStack) GetImage(id, dirName string) (*os.File, error) {
	file, err := os.Create(filepath.Join(dirName, id))
	if err != nil {
		return nil, err
	}

	downloader := s3manager.NewDownloader(l.Session)

	l.logger.Infof("ID:%s", id)

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String("images"),
			Key:    aws.String(id),
		})
	if err != nil {
		l.logger.Infof("error while downloading:%s", err)
		return nil, err
	}

	l.logger.Info("Downloaded ", file.Name(), " ", numBytes, " bytes")

	return file, nil
}

func (l *LocalStack) UploadFiles(images map[string]*os.File) (map[string]string, error) {
	uploader := s3manager.NewUploader(l.Session)

	var results = make(map[string]string)

	for i, file := range images {
		openedFile, _ := os.Open(file.Name())
		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(l.Bucket),
			Key:    aws.String(file.Name()),
			Body:   openedFile,
		})

		err = openedFile.Close()
		if err != nil {
			l.logger.Info(err)
			return nil, err
		}

		if err != nil {
			l.logger.Info(err)
			return nil, err
		}

		l.logger.Infof("File uploaded to %s", result.Location)
		results[i] = result.Location
	}

	return results, nil
}
