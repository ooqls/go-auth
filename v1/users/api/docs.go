package usersapi

import (
	"embed"
	"io/fs"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/go-auth/v1/swaggerui"
)

//go:embed docs/openapi.yaml docs/schemas.yaml
var docsFS embed.FS

var openapiFS, _ = fs.Sub(docsFS, "docs")

func RegisterDocsHandler(e gin.IRouter) {
	swaggerui.RegisterDocsHandler(e, "users API", openapiFS)
}
