package main

import (
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"

	"server/internal/handler"
	"server/internal/localstack"
	"server/internal/service"
)

const (
	bucketName = "images"
)

func main() {

	logger := logrus.New()

	writer := &kafka.Writer{
		Addr:     kafka.TCP("kafka_first", "kafka_second", "kafka_third"),
		Topic:    "images",
		Balancer: &kafka.LeastBytes{},
	}

	localStack, err := localstack.GetConnection(logger, bucketName)
	if err != nil {
		logger.Fatalf("Localstack error: %s", err)
	}

	err = localStack.CreateBucket()
	if err != nil {
		logger.Fatalf("Create bucker error:%s", err)
	}

	resizerService := service.NewService(logger, writer, localStack)
	kafkaReader := resizerService.StartReader()
	resizerService.KafkaReader = kafkaReader

	resizerHandler := handler.NewHandler(resizerService)

	server := gin.New()
	handler.Mount(server, resizerHandler)
	server.Run(":8086")

}
