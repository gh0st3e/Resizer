package service

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"os"
	"resizer/internal/localstack"
	img "resizer/internal/service/image"
	"strings"
	"time"
)

const (
	Small  = "small"
	Medium = "medium"
	Large  = "large"

	path = "images/"
)

type Service struct {
	logger     *logrus.Logger
	localstack *localstack.LocalStack
	kafka      *kafka.Writer
	image      *img.ImgService
}

func NewService(logger *logrus.Logger, localstack *localstack.LocalStack, kafka *kafka.Writer, image *img.ImgService) *Service {
	return &Service{
		logger:     logger,
		localstack: localstack,
		kafka:      kafka,
		image:      image,
	}
}

func (s *Service) StartReader() {
	conf := kafka.ReaderConfig{
		Brokers:        []string{"kafka_first", "kafka_second", "kafka_third"},
		Topic:          "images",
		GroupID:        "gr1",
		MaxBytes:       500,
		CommitInterval: time.Second,
	}

	reader := kafka.NewReader(conf)

	s.logger.Info("Kafka started...")

	for {

		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			s.logger.Infof("(Kafka) Error getting message: %s", err)
			continue
		}
		s.logger.Infof("Message: %s", string(m.Value))

		go s.StartResizeProcess(string(m.Value))
	}
}

func (s *Service) StartResizeProcess(msg string) {
	dirName := s.SetDirName(msg)
	err := os.Mkdir(dirName, 0750)
	if err != nil {
		s.logger.Infof("error creating dir:%s", err)

	}
	defer func() {
		err := os.RemoveAll(dirName)
		if err != nil {
			s.logger.Infof("error deleteing dir:%s", err)
		}
	}()

	file, err := s.localstack.GetImage(msg, dirName)
	if err != nil {
		s.logger.Infof("error getting image:%s", err)

	}
	s.logger.Infof("Get File:%s", file.Name())

	images, err := s.ResizeImg(file, msg, dirName)
	if err != nil {
		s.logger.Infof("Couldn't resize image:%s", err)

	}

	results, err := s.localstack.UploadFiles(images)
	if err != nil {
		s.logger.Infof("Couldn't upload images:%s", err)

	}

	err = s.SendMsg(results)
	if err != nil {
		s.logger.Infof("(Kafka) Error sending message: %s", err)

	}

}

func (s *Service) ResizeImg(file *os.File, id, dirName string) (map[string]*os.File, error) {
	images, err := s.image.ResizeImage(file, id, dirName)
	if err != nil {
		return nil, err
	}

	err = file.Close()
	if err != nil {
		s.logger.Info(err)
		return nil, err
	}

	return images, nil
}

func (s *Service) SendMsg(results map[string]string) error {
	imageStruct := Images{
		Small:  results[Small],
		Medium: results[Medium],
		Large:  results[Large],
	}

	jsonIS, err := json.Marshal(imageStruct)

	s.logger.Info(string(jsonIS))

	err = s.kafka.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte("Image"),
			Value: []byte(jsonIS),
		})
	if err != nil {
		s.logger.Infof("(Kafka) Error sending data: %s", err)
		return err
	}
	s.logger.Info("Message was sent")
	return nil
}

func (s *Service) SetDirName(msg string) string {
	return strings.SplitAfterN(msg, "-", 2)[0]

}
