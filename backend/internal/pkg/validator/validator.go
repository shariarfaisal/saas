package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Validate is the shared validator instance with custom rules registered.
var Validate *validator.Validate

func init() {
	Validate = validator.New()

	// Register custom validators
	Validate.RegisterValidation("phone_bd", validateBDPhone)
	Validate.RegisterValidation("uuid", validateUUID)
	Validate.RegisterValidation("decimal", validateDecimal)
}

// bdPhoneRegex matches Bangladesh phone numbers: +880XXXXXXXXXX or 01XXXXXXXXX
var bdPhoneRegex = regexp.MustCompile(`^(\+880|0)1[3-9]\d{8}$`)

// decimalRegex matches decimal strings like "150.00", "0.5", "1000"
var decimalRegex = regexp.MustCompile(`^\d+(\.\d{1,4})?$`)

func validateBDPhone(fl validator.FieldLevel) bool {
	return bdPhoneRegex.MatchString(fl.Field().String())
}

func validateUUID(fl validator.FieldLevel) bool {
	_, err := uuid.Parse(fl.Field().String())
	return err == nil
}

func validateDecimal(fl validator.FieldLevel) bool {
	return decimalRegex.MatchString(fl.Field().String())
}

// FormatErrors converts validator errors into a map suitable for error responses.
func FormatErrors(err error) map[string]interface{} {
	errors := make(map[string]interface{})
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors[e.Field()] = map[string]string{
				"tag":     e.Tag(),
				"value":   e.Param(),
				"message": formatMessage(e),
			}
		}
	}
	return errors
}

func formatMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "email":
		return e.Field() + " must be a valid email"
	case "phone_bd":
		return e.Field() + " must be a valid Bangladesh phone number"
	case "uuid":
		return e.Field() + " must be a valid UUID"
	case "decimal":
		return e.Field() + " must be a valid decimal number"
	case "min":
		return e.Field() + " must be at least " + e.Param()
	case "max":
		return e.Field() + " must be at most " + e.Param()
	default:
		return e.Field() + " failed validation: " + e.Tag()
	}
}
