package service

import (
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"server/internal/localstack"
	"time"
)

type Service struct {
	Logger      *logrus.Logger
	Kafka       *kafka.Writer
	LocalStack  *localstack.LocalStack
	KafkaReader *kafka.Reader
}

func NewService(logger *logrus.Logger, kafka *kafka.Writer, localStack *localstack.LocalStack) *Service {
	return &Service{
		Logger:     logger,
		Kafka:      kafka,
		LocalStack: localStack,
	}
}

// SendMsg func allows to send msg to 2nd microservice through Kafka
func (s *Service) SendMsg(id string) error {
	err := s.Kafka.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte("Image"),
			Value: []byte(id),
		})
	if err != nil {
		return err
	}
	s.Logger.Infof("ID (%s) was sent", id)

	return nil
}

func (s *Service) StartReader() *kafka.Reader {
	conf := kafka.ReaderConfig{
		Brokers:        []string{"kafka_first", "kafka_second", "kafka_third"},
		Topic:          "messages",
		GroupID:        "gr1",
		MaxBytes:       500,
		CommitInterval: time.Second,
	}

	reader := kafka.NewReader(conf)

	return reader

}

func (s *Service) GetMessage() (string, error) {
	for {
		m, err := s.KafkaReader.ReadMessage(context.Background())
		if err != nil {
			s.Logger.Info(err)
			return "", err
		}
		if len(string(m.Value)) > 1 {
			return string(m.Value), nil
		}
	}
}
