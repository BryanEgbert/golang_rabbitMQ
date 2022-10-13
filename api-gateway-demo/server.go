package main

import (
	handler "api-gateway-demo/handlers"
	"log"
	"net/http"
	"strings"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/streadway/amqp"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, strings.Split(err.Error(), ":")[2])
	}

	return nil
}

func main() {
	// Connect to rabbitMQ server
	amqpServerURL := "amqp://guest:guest@localhost:5672/"

	rabbitMq, err := amqp.Dial(amqpServerURL)
	if err != nil {
		panic(err)
	}
	defer rabbitMq.Close()

	channelRabbitMQ, err := rabbitMq.Channel()
	if err != nil {
		panic(err)
	}
	defer channelRabbitMQ.Close()

	_, err = channelRabbitMQ.QueueDeclare(
		"user_queue",
		true,
		false,
		false,
		false,
		nil,
	)

	replyQueue, err := channelRabbitMQ.QueueDeclare(
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	log.Print(replyQueue.Name)

	if err != nil {
		panic(err)
	}

	messages, err := channelRabbitMQ.Consume(
		replyQueue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	server := make(chan bool)

	go func() {
		// Create a server
		e := echo.New()

		e.Use(middleware.Logger())
		e.Use(middleware.Recover())
		e.Use(middleware.Gzip())

		e.Validator = &CustomValidator{validator: validator.New()}

		var handler handler.IUserHandler = handler.NewHandler(channelRabbitMQ, replyQueue.Name)

		e.GET("/api-gateway/user", handler.GetUserData)
		e.GET("/api-gateway/queue", handler.SimulateMQ)

		e.Logger.Fatal(e.Start(":8000"))
	}()

	receiver := make(chan bool)

	go func() {
		for message := range messages {
			log.Printf("\tReceived message: %s\n", message.Body)
			log.Printf("\tCorrelation ID: %s\n", message.CorrelationId)
		}
	}()

	<-server
	<-receiver
}
