package helper

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})
	return v
}

func ValidateStruct(s any) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return map[string]string{"_": "validasi gagal"}
	}

	result := make(map[string]string, len(fieldErrors))
	for _, fe := range fieldErrors {
		result[fe.Field()] = messageFor(fe)
	}
	return result
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "min":
		if fe.Kind() == reflect.String {
			return "minimal " + fe.Param() + " karakter"
		}
		if fe.Kind() == reflect.Slice || fe.Kind() == reflect.Array {
			return "minimal " + fe.Param() + " item" // Perbaikan dari modul
		}
		return "nilai minimal " + fe.Param()
	case "max":
		if fe.Kind() == reflect.String {
			return "maksimal " + fe.Param() + " karakter"
		}
		if fe.Kind() == reflect.Slice || fe.Kind() == reflect.Array {
			return "maksimal " + fe.Param() + " item" // Perbaikan dari modul
		}
		return "nilai maksimal " + fe.Param()
	default:
		return "tidak memenuhi aturan " + fe.Tag()
	}
}
