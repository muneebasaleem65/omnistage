package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/muneebasaleem65/omnistage/internal/storage"
	"github.com/muneebasaleem65/omnistage/internal/utils/response"
)

var validate = validator.New()

type eventRequest struct {
	VenueID     int64     `json:"venue_id" asdf1230validate:"required"`
	Title       string    `json:"title" validate:"required"`
	Artist      string    `json:"artist" validate:"required"`
	Description string    `json:"description" validate:"required"`
	StartsAt    time.Time `json:"starts_at" validate:"required,gt"`
}

func Create(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req eventRequest

		err := json.NewDecoder(r.Body).Decode(&req)

		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		if err := validate.Struct(req); err != nil {
			var validateErrs validator.ValidationErrors
			if errors.As(err, &validateErrs) {
				response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
				return
			}
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		lastID, err := store.CreateEvent(
			req.VenueID,
			req.Title,
			req.Artist,
			req.Description,
			req.StartsAt,
		)

		if err != nil {
			if errors.Is(err, storage.ErrVenueNotFound) {
				response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("venue not found")))
				return
			}
			slog.Error("failed to create an event", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("internal server error")))
			return
		}

		response.WriteJson(w, http.StatusCreated, map[string]any{"id": lastID})
	}
}

func Index(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		city := q.Get("city")

		limit := 20
		if v:= q.Get("limit"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n <= 0 || n > 100 {
				response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("invalid limit")))
				return
			}
			limit = n
		}

		offset := 0
		if v := q.Get("offset"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				response.WriteJson(w, http.StatusBadRequest,
					response.GeneralError(fmt.Errorf("invalid offset")))
				return
			}
			offset = n
		}

		var from, to *time.Time
		if v := q.Get("from"); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				response.WriteJson(w, http.StatusBadRequest,
					response.GeneralError(fmt.Errorf("invalid from date")))
				return
			}
			from = &t
		}

		if v := q.Get("to"); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				response.WriteJson(w, http.StatusBadRequest,
					response.GeneralError(fmt.Errorf("invalid to date")))
				return
			}
			to = &t
		}

		if from != nil && to != nil && to.Before(*from) {
			response.WriteJson(w, http.StatusBadRequest,
				response.GeneralError(fmt.Errorf("'to' date cannot be before 'from' date")))
			return
		}

		events, total, err := store.ListEvents(city, from, to, limit, offset)
		if err != nil {
			slog.Error("failed to list events", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("internal error")))
			return
		}
		response.WriteJson(w, http.StatusOK, map[string]any{
			"data": events,
			"meta": map[string]any{
				"total":  total,
				"limit":  limit,
				"offset": offset,
			},
		})
	}
}


