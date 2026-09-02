package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/muneebasaleem65/omnistage/internal/types"
	"github.com/muneebasaleem65/omnistage/internal/utils/response"
)

func New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user user.User

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

		if err := validator.New().Struct(user); err != nil {
			//typecasting the err to use in ValidationError() function

			validateErrs := err.(validator.ValidationErrors)

			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
			return
		}

		response.WriteJson(w, http.StatusCreated, map[string]string{"success": "OK"})
	}
}
