package localstack

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"mime/multipart"
	"os"
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

func (l *LocalStack) CreateBucket() error {
	bucket, err := l.Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(l.Bucket),
	})
	if err != nil {
		l.logger.Infof("Creating bucket error: %s", err)
		return err
	}
	l.logger.Infof("Bucket was successfuly created: %s", bucket)
	return nil
}

func (l *LocalStack) SetData(id string, file multipart.File, fileHeader *multipart.FileHeader) error {

	size := fileHeader.Size
	buffer := make([]byte, size)
	file.Read(buffer)

	l.logger.Info(buffer)

	//tempFileName := "images/" + bson.NewObjectId().Hex() + filepath.Ext(fileHeader.Filename)

	object, err := l.Client.PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(l.Bucket),
		Key:                aws.String(id),
		ACL:                aws.String("public-read"),
		Body:               bytes.NewReader([]byte("test")),
		ContentLength:      aws.Int64(int64(4)),
		ContentType:        aws.String("application/text"),
		ContentDisposition: aws.String("attachment"),
	})
	if err != nil {
		l.logger.Infof("Couldn't set data: %s", err)
		return err
	}
	l.logger.Infof("Successfuly written: %v", object)
	return nil
}

func (l *LocalStack) GetBuckets() {
	result, _ := l.Client.ListBuckets(&s3.ListBucketsInput{})
	l.logger.Info("Buckets:")
	for _, bucket := range result.Buckets {
		l.logger.Info(*bucket.Name + ": " + bucket.CreationDate.Format("2006-01-02 15:04:05 Monday"))
	}
}

func (l *LocalStack) UploadFile(file multipart.File, fileName string) (string, error) {
	//file, err := header.Open()
	//if err != nil {
	//	return "", err
	//}
	uploader := s3manager.NewUploader(l.Session)

	id, err := uuid.NewUUID()
	if err != nil {
		l.logger.Infof("Couldn't create uuid:%s", err)
		return "", err
	}

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(l.Bucket),
		Key:    aws.String(id.String() + "_" + fileName),
		Body:   file,
	})
	if err != nil {
		l.logger.Info(err)
	}
	l.logger.Infof("File uploaded to %s", result.Location)

	return id.String() + "_" + fileName, nil
}

func (l *LocalStack) GetImage() error {

	file, err := os.Create("downloaded_image.png")
	if err != nil {
		fmt.Println(err)
	}

	defer file.Close()

	downloader := s3manager.NewDownloader(l.Session)

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(l.Bucket),
			Key:    aws.String("863c3f3d-9348-11ed-9f56-e8d8d1f76e0b_images.png"),
		})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

	return nil
}
