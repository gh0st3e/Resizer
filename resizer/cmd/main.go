package main

import (
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"resizer/internal/localstack"
	"resizer/internal/service"
	"resizer/internal/service/image"
)

const (
	bucketName = "images"
)

func main() {
	logger := logrus.New()

	writer := &kafka.Writer{
		Addr:     kafka.TCP("kafka_first", "kafka_second", "kafka_third"),
		Topic:    "messages",
		Balancer: &kafka.LeastBytes{},
	}

	localStack, err := localstack.GetConnection(logger, bucketName)
	if err != nil {
		logger.Fatalf("Localstack error: %s", err)
	}

	imageService := image.NewImage(logger)

	resizerService := service.NewService(logger, localStack, writer, imageService)

	resizerService.StartReader()
}
