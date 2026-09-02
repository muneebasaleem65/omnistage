package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/muneebasaleem65/omnistage/internal/storage"
	"github.com/muneebasaleem65/omnistage/internal/types"
	"github.com/muneebasaleem65/omnistage/internal/utils/response"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

func Register(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user types.User

		err := json.NewDecoder(r.Body).Decode(&user)

		//if body is empty
		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		//request validation

		if err := validate.Struct(user); err != nil {
			var validateErrs validator.ValidationErrors
			if errors.As(err, &validateErrs) {
				response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
				return
			}
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			slog.Error("failed to hash password", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		lastId, err := store.CreateUser(
			user.Email,
			string(bcryptPassword),
			user.Name,
		)
		if err != nil {
			if errors.Is(err, storage.ErrUserExists) {
				response.WriteJson(w, http.StatusConflict, response.GeneralError(fmt.Errorf("email already registered")))
				return
			}
			slog.Error("failed to create user", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("internal error")))
			return
		}

		response.WriteJson(w, http.StatusCreated, map[string]any{"id": lastId})
	}
}
