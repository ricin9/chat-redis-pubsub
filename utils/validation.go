package utils

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

func FormatErrors(err error) map[string]string {
	errors := make(map[string]string)

	for _, err := range err.(validator.ValidationErrors) {
		switch err.Tag() {
		case "required":
			errors[err.Field()] = fmt.Sprintf("%s is required", err.Field())
		case "min":
			errors[err.Field()] = fmt.Sprintf("%s must be at least %s characters long", err.Field(), err.Param())
		case "max":
			errors[err.Field()] = fmt.Sprintf("%s must not exceed %s characters", err.Field(), err.Param())
		case "email":
			errors[err.Field()] = fmt.Sprintf("%s must be a valid email address", err.Field())
		case "gte":
			errors[err.Field()] = fmt.Sprintf("%s must be greater than or equal to %s", err.Field(), err.Param())
		case "lte":
			errors[err.Field()] = fmt.Sprintf("%s must be less than or equal to %s", err.Field(), err.Param())
		case "containsany":
			errors[err.Field()] = fmt.Sprintf("%s must contain at least one of the following: %s", err.Field(), err.Param())
		default:
			errors[err.Field()] = fmt.Sprintf("%s failed on the '%s' tag", err.Field(), err.Param())
		}
	}

	return errors
}
