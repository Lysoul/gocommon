package shared

import (
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
)

//nolint:gochecknoglobals // sync.OnceValue is safe for concurrent use by multiple goroutines.
var Validator = sync.OnceValue(func() *validator.Validate {
	return validator.New()
})

// todo investigate using json schema for validation
// https://github.com/santhosh-tekuri/jsonschema
func MapValidationErrors(errs *validator.ValidationErrors) []string {
	details := make([]string, len(*errs))
	for errNum, err := range *errs {
		path := strings.Split(err.Namespace(), ".")
		field := make([]string, len(path)-1)
		for i, p := range path[1:] {
			field[i] += strcase.ToLowerCamel(p)
		}
		details[errNum] = strings.Join(field, ".") + " " + translateTag(err)
	}

	return details
}

func translateTag(err validator.FieldError) string {
	switch err.ActualTag() {
	case "gtfield":
		return "must greater than " + strcase.ToLowerCamel(err.Param())
	case "gtefield":
		return "must be greater than or equal " + strcase.ToLowerCamel(err.Param())
	case "lte":
		return "must be maximum " + err.Param()
	case "gt":
		return "must be greater than " + err.Param()
	case "gte":
		return "must be greater than or equal " + err.Param()
	case "ltefield":
		return "must be less then or equal " + strcase.ToLowerCamel(err.Param())
	case "required_with":
		return "required with " + strcase.ToLowerCamel(err.Param())
	case "oneof":
		return "must be one of " + err.Param()
	default:
		return err.ActualTag()
	}
}
