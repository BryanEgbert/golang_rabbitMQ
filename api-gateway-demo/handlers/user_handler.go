package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"
)

type (
	IUserHandler interface {
		GetUserData(c echo.Context) error
		SimulateMQ(c echo.Context) error
	}

	User struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	Handler struct {
		channelRabbitMQ *amqp.Channel
	}
)

func NewHandler(channel *amqp.Channel) *Handler {
	return &Handler{
		channelRabbitMQ: channel,
	}
}

func (h *Handler) GetUserData(c echo.Context) error {
	data, err := http.Get("http://localhost:8080/api/v1")

	var userData User

	if err != nil {
		c.Logger().Panic(err.Error())
	}
	defer data.Body.Close()

	body, err := io.ReadAll(data.Body)
	if err != nil {
		c.Logger().Panic(err.Error())
	}
	json.Unmarshal(body, &userData)

	return c.JSON(http.StatusOK, userData)
}

func (h *Handler) SimulateMQ(c echo.Context) error {
	message := amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(`{"name": "Bob", "email": "test@test.com"}`),
	}

	if err := h.channelRabbitMQ.Publish(
		"",
		"user_queue",
		false,
		false,
		message,
	); err != nil {
		return err
	}

	c.Logger().Debug(message)

	return nil
}
