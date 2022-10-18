package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

var MessageChan chan string

func prepareHeaderForSSE(c echo.Context) {
	c.Request().Header.Add("Content-Type", "text/event-stream")
	c.Request().Header.Add("Cache-Control", "no-cache")
	c.Request().Header.Add("Connection", "keep-alive")
	c.Request().Header.Add("Access-Control-Allow-Origin", "*")
}

func SseStream(c echo.Context) error {
	log.Print("Connected to event")
	w := c.Response().Writer

	prepareHeaderForSSE(c)
	MessageChan = make(chan string)
	flusher, _ := w.(http.Flusher)

	for {
		select {
		case message := <-MessageChan:
			log.Print(message)
			fmt.Fprintf(w, "%s\n", message)
			flusher.Flush()
		}
	}
}

func SseMessage(message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if MessageChan != nil {
			MessageChan <- message
		}
	}
}
