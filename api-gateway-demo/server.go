package main

import (
	handler "api-gateway-demo/handlers"
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

	if err != nil {
		panic(err)
	}

	// Create a server
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	e.Validator = &CustomValidator{validator: validator.New()}

	var handler handler.IUserHandler = handler.NewHandler(channelRabbitMQ)

	e.GET("/api-gateway/user", handler.GetUserData)
	e.GET("/api-gateway/queue", handler.SimulateMQ)

	e.Logger.Fatal(e.Start(":8000"))
}
