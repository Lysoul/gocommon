package shared

import (
	"errors"
	"fmt"
	"strings"

	// "github.com/Lysoul/gocommon/monitoring"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var (
	ErrUnauthorized     = ConstError("unauthorized")
	ErrPermissionDenied = ConstError("permission_denied")
	ErrNotFound         = ConstError("not_found")
	ErrValidationFailed = ConstError("validation_failed")
	ErrUnexpected       = ConstError("unexpected")
)

type ConstError string

func (err ConstError) Error() string {
	return string(err)
}
func (err ConstError) Is(target error) bool {
	ts := target.Error()
	es := string(err)
	return ts == es || strings.HasPrefix(ts, es+": ")
}
func (err ConstError) Wrap(inner error) error {
	return wrapError{msg: string(err), err: inner}
}

type wrapError struct {
	err error
	msg string
}

func (err wrapError) Error() string {
	if err.err != nil {
		return fmt.Sprintf("%s: %v", err.msg, err.err)
	}
	return err.msg
}
func (err wrapError) Unwrap() error {
	return err.err
}
func (err wrapError) Is(target error) bool {
	return ConstError(err.msg).Is(target)
}

func ErrorToHTTP(ctx *gin.Context, err error) {
	validationErrs := &validator.ValidationErrors{}
	switch {
	case err == nil:
		return
	case errors.As(err, validationErrs):
		ctx.AbortWithStatusJSON(400, gin.H{
			"error":   "validation_failed",
			"details": MapValidationErrors(validationErrs),
		})
	case errors.Is(err, ErrNotFound),
		errors.Is(err, ErrPermissionDenied):
		ctx.AbortWithStatus(404)
	default:
		// monitoring.Logger().ErrorContext(ctx, "no http mapping for error",
		// 	zap.String("type", fmt.Sprintf("%T", err)),
		// 	zap.Error(err))

		ctx.AbortWithStatus(500)
	}
	// todo: only handle errors we know and call shared.handleError(err) for the rest
}
