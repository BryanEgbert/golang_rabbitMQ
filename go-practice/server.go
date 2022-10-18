package main

import (
	"log"
	"myapp/handler"
	service "myapp/services"
	"time"

	"net/http"
	"strings"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/streadway/amqp"
	"github.com/uptrace/bun/driver/sqliteshim"
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

func consumeQueue() {
	amqpServerURL := "amqp://guest:guest@localhost:5672/"
	connectRabbitMQ, err := amqp.Dial(amqpServerURL)
	if err != nil {
		panic(err)
	}

	defer connectRabbitMQ.Close()

	channel, err := connectRabbitMQ.Channel()
	if err != nil {
		panic(err)
	}

	defer channel.Close()

	messages, err := channel.Consume(
		"user_queue",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println(err)
	}

	receiver := make(chan bool)

	go func() {
		for message := range messages {
			log.Print("Processing...")
			time.Sleep(5 * time.Second)
			body := amqp.Publishing{
				ContentType:   "application/json",
				Body:          []byte(`{"name": "Bob", "email": "bob123@test.com"}`),
				CorrelationId: message.CorrelationId,
			}
			log.Printf("\tReceived message: %s\n", message.Body)
			log.Printf("\tCorrelation ID: %s\n", message.CorrelationId)
			log.Printf("\tReply to: %s\n", message.ReplyTo)
			channel.Publish("", message.ReplyTo, true, false, body)
			channel.Ack(message.DeliveryTag, true)
			log.Print("Done...")
		}
	}()

	<-receiver
}

func main() {
	// Code for http request/response

	server := make(chan bool)

	go func() {
		e := echo.New()

		e.Use(middleware.Recover())

		e.Validator = &CustomValidator{validator: validator.New()}

		var db service.DB = handler.CreateDB(sqliteshim.ShimName, "file::memory:")

		sqldb, err := db.Connect()
		if err != nil {
			panic(err)
		}

		var userHandler handler.UserInterface = handler.CreateUserHandler(sqldb)
		var HealthCheckHandler handler.HealthCheckInterface = handler.CreateHealthCheckHandler(db)

		e.GET("/api/v1", userHandler.GreetUser)
		e.POST("/api/v1/token", userHandler.GenerateToken)
		e.GET("/api/v1/health", HealthCheckHandler.GetHealthCheck)

		e.Logger.Fatal(e.Start(":8080"))
	}()

	consumeQueue()
	<-server
}
