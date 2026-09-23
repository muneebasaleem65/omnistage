package venue

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/muneebasaleem65/omnistage/internal/storage"
	"github.com/muneebasaleem65/omnistage/internal/utils/response"
)

// creating a new validator
var validate = validator.New()

// validating the data
type venueRequest struct {
	Name     string `json:"name" validate:"required"`
	City     string `json:"city" validate:"required"`
	Capacity int    `json:"capacity" validate:"required,gt=0"`
}

func Create(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req venueRequest

		err := json.NewDecoder(r.Body).Decode(&req)

		//if body is empty
		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		//validation errors
		//here we are using the struct method on validate variable that
		//holds the new validator the req variable holds the venueRequest struct
		if err := validate.Struct(req); err != nil {
			var validateErrs validator.ValidationErrors
			if errors.As(err, &validateErrs) {
				response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
				return
			}
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		lastID, err := store.CreateVenue(
			req.Name,
			req.City,
			req.Capacity,
		)

		if err != nil {
			slog.Error("failed to create a venue", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("internal server error")))
			return
		}

		response.WriteJson(w, http.StatusCreated, map[string]any{"id": lastID})
	}
}

func Index(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		venues, err := store.ListVenues()
		if err != nil {
			slog.Error("failed to list venues", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("internal error")))
			return
		}

		response.WriteJson(w, http.StatusOK, venues)
	}
}
