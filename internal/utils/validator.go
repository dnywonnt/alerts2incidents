package utils // dnywonnt.me/alerts2incidents/internal/utils

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Global validator instance
var validate *validator.Validate

// init initializes the global validator
func init() {
	validate = validator.New()
	validate.RegisterValidation("required_if_m", requiredIfM)
}

// Custom validation function to check if the field value is required based on another field's value
func requiredIfM(fl validator.FieldLevel) bool {
	param := fl.Param()
	params := strings.Split(param, " ")

	if len(params) < 2 {
		return false
	}

	fieldName := params[0]
	allowedValues := params[1:]

	// Get the struct value
	structValue := fl.Top()

	// Find the field by name
	field := reflect.Indirect(structValue).FieldByName(fieldName)
	if !field.IsValid() {
		return false
	}

	fieldValue := field.String()

	// Check if the field value matches one of the allowed values
	for _, value := range allowedValues {
		if fieldValue == value {
			// If the value matches, the current field must not be empty
			return fl.Field().String() != ""
		}
	}

	return true
}

// ValidateStruct validates a structure using the global validator
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}
