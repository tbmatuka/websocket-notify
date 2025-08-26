package websocket_notify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Event struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
	Data string   `json:"data"`
}

func handleEventRequest(responseWriter http.ResponseWriter, request *http.Request, config Config) {
	if request.Method != http.MethodPost {
		http.Error(responseWriter, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	requestAuth := request.Header.Get(`Auth`)
	if requestAuth != config.ApiSecret {
		http.Error(responseWriter, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

		return
	}

	request.Body = http.MaxBytesReader(responseWriter, request.Body, 1024*1024)

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()

	var event Event
	err := decoder.Decode(&event)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf(`Request body contains badly-formed JSON (at position %d)`, syntaxError.Offset)
			http.Error(responseWriter, msg, http.StatusBadRequest)
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := `Request body contains badly-formed JSON`
			http.Error(responseWriter, msg, http.StatusBadRequest)
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf(`Request body contains an invalid value for the %q field (at position %d)`, unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(responseWriter, msg, http.StatusBadRequest)
		case strings.HasPrefix(err.Error(), `json: unknown field `):
			fieldName := strings.TrimPrefix(err.Error(), `json: unknown field `)
			msg := `Request body contains unknown field ` + fieldName
			http.Error(responseWriter, msg, http.StatusBadRequest)
		case errors.Is(err, io.EOF):
			msg := `Request body must not be empty`
			http.Error(responseWriter, msg, http.StatusBadRequest)
		case err.Error() == `http: request body too large`:
			msg := `Request body must not be larger than 1MB`
			http.Error(responseWriter, msg, http.StatusRequestEntityTooLarge)
		default:
			log.Print(err.Error())
			http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		return
	}

	logger.Debug(fmt.Sprintf(`API event: %s -> %s`, event.Name, strings.Join(event.Tags, `, `)))

	manager := getSubscriptionManager()
	manager.DistributeEvent(event)
}

type ApiStatus struct {
	Connections   uint            `json:"connections"`
	Subscriptions map[string]uint `json:"subscriptions"`
}

func handleStatusRequest(responseWriter http.ResponseWriter, request *http.Request, config Config) {
	if request.Method != http.MethodGet {
		http.Error(responseWriter, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

		return
	}

	requestAuth := request.Header.Get(`Auth`)
	if requestAuth != config.ApiSecret {
		http.Error(responseWriter, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

		return
	}

	manager := getSubscriptionManager()
	status := manager.Status()

	response, err := json.Marshal(status)
	if err != nil {
		http.Error(responseWriter, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	responseWriter.WriteHeader(http.StatusOK)
	responseWriter.Header().Set("Content-Type", "application/json")
	_, _ = responseWriter.Write(response)
}
