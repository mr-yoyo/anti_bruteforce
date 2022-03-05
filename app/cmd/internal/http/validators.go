package http

import (
	"errors"
	"net"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func registerValidations() {
	_ = validate.RegisterValidation("ip4_net", func(fl validator.FieldLevel) bool {
		val, ok := fl.Field().Interface().(string)
		if !ok {
			return false
		}

		_, ip4Net, err := net.ParseCIDR(val)
		if err != nil {
			return false
		}

		if len(ip4Net.Mask) != 4 {
			return false
		}

		return true
	})
}

func validationFailed(w http.ResponseWriter, err error) {
	r := make([]string, 0)
	var valErrs validator.ValidationErrors
	if errors.As(err, &valErrs) {
		for _, err := range valErrs {
			r = append(r, err.Error())
		}
	} else {
		r = append(r, err.Error())
	}

	bytes, _ := json.Marshal(struct {
		Errors []string `json:"errors"`
	}{
		r,
	})
	http.Error(w, string(bytes), http.StatusBadRequest)
}
