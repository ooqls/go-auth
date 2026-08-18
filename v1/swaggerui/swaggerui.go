package swaggerui

import (
	"bytes"
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path"
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
// openapiFS must contain openapi.yaml at its root. Any other files at the root
// (e.g. schemas.yaml referenced via external $ref) are served alongside it so
// Swagger UI can resolve those references.
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

	// Serve every file at the root of openapiFS (openapi.yaml plus any
	// externally-referenced specs such as schemas.yaml).
	entries, err := fs.ReadDir(openapiFS, ".")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		contentType := mime.TypeByExtension(path.Ext(name))
		switch path.Ext(name) {
		case ".yaml", ".yml":
			contentType = "application/yaml"
		}
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		e.GET("/docs/"+name, func(c *gin.Context) {
			f, err := openapiFS.Open(name)
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
			defer f.Close()
			c.DataFromReader(http.StatusOK, -1, contentType, f, nil)
		})
	}

	e.StaticFS("/docs/assets", http.FS(assets))

	e.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "docs/")
	})
}
