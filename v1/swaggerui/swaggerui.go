package swaggerui

import (
	"bytes"
	"embed"
	"io/fs"
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"
)

//go:embed assets
var assetsFS embed.FS

var indexTmpl = template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html>
<head>
  <title>{{.Title}}</title>
  <meta charset="UTF-8">
  <link rel="stylesheet" href="assets/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="assets/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: "openapi.yaml",
      dom_id: '#swagger-ui',
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
      layout: "BaseLayout"
    })
  </script>
</body>
</html>
`))

// RegisterDocsHandler mounts a Swagger UI at /docs/ on the given router.
// title is shown in the browser tab.
// openapiFS must contain openapi.yaml at its root.
func RegisterDocsHandler(e gin.IRouter, title string, openapiFS fs.FS) {
	var buf bytes.Buffer
	if err := indexTmpl.Execute(&buf, struct{ Title string }{title}); err != nil {
		panic(err)
	}
	indexHTML := buf.Bytes()

	assets, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		panic(err)
	}

	e.GET("/docs/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	e.GET("/docs/openapi.yaml", func(c *gin.Context) {
		f, err := openapiFS.Open("openapi.yaml")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		defer f.Close()
		c.DataFromReader(http.StatusOK, -1, "application/yaml", f, nil)
	})

	e.StaticFS("/docs/assets", http.FS(assets))

	e.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "docs/")
	})
}
