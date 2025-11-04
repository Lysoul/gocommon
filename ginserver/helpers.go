package ginserver

import (
	"slices"

	"github.com/gin-gonic/gin"
)

// Gets accepted language for request header accept-language
// if the language is not in `supported` list, defaults to first supported language.
func GetAcceptedLanguage(ctx *gin.Context, supported []string) string {
	lang := ctx.GetHeader("accept-language")
	if slices.Contains(supported, lang) {
		return lang
	} else {
		return supported[0]
	}
}
