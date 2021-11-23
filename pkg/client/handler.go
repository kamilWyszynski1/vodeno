package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
	"github.com/sirupsen/logrus"
)

// Handler is a http handler for clients.
type Handler struct {
	service   Service
	validator *validator.Validate
	log       *logrus.Logger
}

// NewHandler returns new instance of Handler.
func NewHandler(log *logrus.Logger, svc Service) *Handler {
	return &Handler{
		service:   svc,
		validator: validator.New(),
		log:       log,
	}
}

// AddRoutes adds client routes to router.
func (h *Handler) AddRoutes(router chi.Router) {
	router.Route("/clients", func(r chi.Router) {
		r.Post("/", h.add)
		r.Post("/send", h.send)
		r.Delete("/{id}", h.delete)
		r.Get("/", h.list)
		r.Get("/{id}", h.get)
	})
}

// add gets Entry from http request and calls Service from creation.
func (h *Handler) add(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger := h.log.WithField("handler", "add")
	var req Entry

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("failed to decode request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		logger.WithError(err).Error("request is not valid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.Add(ctx, req); err != nil {
		logger.WithError(err).Error("failed to add client")
		if errors.Is(err, ErrDuplicate) { // handle duplicate error.
			h.writeJSONContentHeader(w)
			w.WriteHeader(http.StatusBadRequest)
			if _, err := w.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err))); err != nil {
				h.log.WithError(err).Error("failed to write error to ResponseWriter")
			}
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// SendRequest is a send handler request.
type SendRequest struct {
	MailingID int `json:"mailing_id" validate:"required"`
}

// send takes SendRequest from http request and call Service's Send method.
func (h *Handler) send(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger := h.log.WithField("handler", "send")
	var req SendRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(err).Error("failed to decode request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	logger = h.log.WithField("mailing_id", req.MailingID)

	if err := h.validator.Struct(req); err != nil {
		logger.WithError(err).Error("request is not valid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.Send(ctx, req.MailingID); err != nil {
		logger.WithError(err).Error("failed to send emails")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// delete is a delete client http handler.
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger := h.log.WithField("handler", "delete")

	clientID := chi.URLParam(r, "id")
	logger = logger.WithField("client_id", clientID)

	if clientID == "" {
		logger.Error("clientID is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(clientID)
	if err != nil {
		logger.Error("clientID must be an integer")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		logger.WithError(err).Error("failed to send emails")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type listResponse struct {
	Clients []Entry `json:"clients"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger := h.log.WithField("handler", "list")

	cursor, err := CursorFromRequest(r)
	if err != nil {
		logger.WithError(err).Error("failed to get cursor")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	clients, err := h.service.List(ctx, *cursor)
	if err != nil {
		logger.WithError(err).Error("failed to get clients")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(clients) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	} else if len(clients) > 0 {
		// write after_id header.
		last := clients[len(clients)-1]
		w.Header().Add("after_id", strconv.Itoa(last.ID))
	}

	h.writeJSONContentHeader(w)
	w.WriteHeader(http.StatusOK)
	// write json.
	if err := json.NewEncoder(w).Encode(listResponse{clients}); err != nil {
		logger.WithError(err).Error("writing JSON to ResponseWriter failed")
	}
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	logger := h.log.WithField("handler", "get")

	clientID := chi.URLParam(r, "id")
	logger = logger.WithField("client_id", clientID)

	if clientID == "" {
		logger.Error("clientID is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(clientID)
	if err != nil {
		logger.Error("clientID must be an integer")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	client, err := h.service.Get(ctx, id)
	if err != nil {
		logger.WithError(err).Error("failed to get client")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if client == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.writeJSONContentHeader(w)
	w.WriteHeader(http.StatusOK)
	// write json.
	if err := json.NewEncoder(w).Encode(client); err != nil {
		logger.WithError(err).Error("writing JSON to ResponseWriter failed")
	}
}

func (h Handler) writeJSONContentHeader(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
}
