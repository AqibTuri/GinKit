// Package validate runs struct tags through go-playground/validator after Gin binding (second validation pass).
package validate

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

var v = validator.New()

// Struct runs go-playground/validator after Gin binding for shared DTO validation.
func Struct(s any) map[string]string {
	err := v.Struct(s)
	if err == nil {
		return nil
	}
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		out := make(map[string]string)
		for _, fe := range verrs {
			out[strings.ToLower(fe.Field())] = fe.Tag()
		}
		return out
	}
	return map[string]string{"_": err.Error()}
}
