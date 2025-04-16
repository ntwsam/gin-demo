package validators

import (
	"github.com/go-playground/validator/v10"
	"github.com/ttacon/libphonenumber"
)

func PhoneValidator(fl validator.FieldLevel) bool {
	phoneStr := fl.Field().String()
	num, err := libphonenumber.Parse(phoneStr, "TH")
	if err != nil {
		return false
	}
	return libphonenumber.IsValidNumber(num)
}
