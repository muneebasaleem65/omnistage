package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/muneebasaleem65/omnistage/internal/http/middleware"
	"github.com/muneebasaleem65/omnistage/internal/storage"
	"github.com/muneebasaleem65/omnistage/internal/token"
	"github.com/muneebasaleem65/omnistage/internal/utils/response"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

type registerRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Name     string `json:"name" validate:"required"`
}

func Register(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerRequest

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

		//request validation

		if err := validate.Struct(req); err != nil {
			var validateErrs validator.ValidationErrors
			if errors.As(err, &validateErrs) {
				response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
				return
			}
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		bcryptPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			slog.Error("failed to hash password", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		lastId, err := store.CreateUser(
			req.Email,
			string(bcryptPassword),
			req.Name,
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

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func Login(store storage.Storage, jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
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
		//request validation
		if err := validate.Struct(req); err != nil {
			var validateErrs validator.ValidationErrors
			if errors.As(err, &validateErrs) {
				response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
				return
			}
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		user, err := store.GetUserByEmail(
			req.Email,
		)
		if err != nil {
			if !errors.Is(err, storage.ErrUserNotFound) {
				slog.Error("failed to get user", slog.String("error", err.Error()))
			}
			response.WriteJson(w, http.StatusUnauthorized, response.GeneralError(fmt.Errorf("invalid email or password")))
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			response.WriteJson(w, http.StatusUnauthorized, response.GeneralError(fmt.Errorf("invalid email or password")))
			return
		}

		signed, err := token.GenerateToken(user.ID, user.Role, jwtSecret)
		if err != nil {
			slog.Error("failed to generate token", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("internal error")))
			return
		}

		response.WriteJson(w, http.StatusOK, map[string]any{"token": signed})
	}
}

func Me(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			response.WriteJson(w, http.StatusUnauthorized,
				response.GeneralError(fmt.Errorf("unauthorized")))
			return
		}

		user, err := store.GetUserByID(userID)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				response.WriteJson(w, http.StatusUnauthorized,
					response.GeneralError(fmt.Errorf("unauthorized")))
				return
			}
			slog.Error("failed to get user", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("internal error")))
			return
		}

		response.WriteJson(w, http.StatusOK, map[string]any{
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		})
	}
}
