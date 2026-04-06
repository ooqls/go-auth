package resourcesapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/resourcesv1"
	"github.com/ooqls/go-auth/v1/resources/resourcesapi/gen_resources"
	"go.uber.org/zap"
)

var _ gen_resources.ServerInterface = &ResourcesServer{}

type ResourcesServer struct {
	resourceReader resourcesv1.Reader
	resourceWriter resourcesv1.Writer
	l              *zap.Logger
}

func NewResourcesServer(factory datav1.Factory, l *zap.Logger) *ResourcesServer {
	return &ResourcesServer{
		resourceReader: factory.NewResourceReader(),
		resourceWriter: factory.NewResourceWriter(),
		l:              l,
	}
}

func (r *ResourcesServer) DeleteResource(c *gin.Context, name string, group string, kind string) {
	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	err := r.resourceWriter.DeleteResource(c, group, kind, name)
	if err != nil {
		if errors.Is(err, resourcesv1.ErrNotFound) {
			r.l.Error("resource not found", zap.Error(err))
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		r.l.Error("failed to delete resource", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Silence unused variable warning
	_ = authCtx

	c.JSON(http.StatusNoContent, nil)
}

func (r *ResourcesServer) GetResource(c *gin.Context, name string, group string, kind string) {
	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	res, err := r.resourceReader.GetResource(c, name, corev1.Metadata{
		Group: group,
		Kind:  kind,
	})
	if err != nil {
		r.l.Error("failed to get resource", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if res == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
		return
	}

	// Silence unused variable warning
	_ = authCtx

	c.JSON(http.StatusOK, res)
}

func (r *ResourcesServer) ListResources(c *gin.Context, params gen_resources.ListResourcesParams) {
	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	group := ""
	kind := ""
	page := int32(0)
	pageSize := int32(100)

	if params.Group != nil {
		group = *params.Group
	}
	if params.Kind != nil {
		kind = *params.Kind
	}
	if params.Page != nil {
		page = int32(*params.Page)
	}
	if params.PageSize != nil {
		pageSize = int32(*params.PageSize)
	}

	res, err := r.resourceReader.GetResources(c, corev1.Metadata{
		Group: group,
		Kind:  kind,
	}, pageSize, page*pageSize)
	if err != nil {
		r.l.Error("failed to get resources", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Silence unused variable warning
	_ = authCtx

	c.JSON(http.StatusOK, res)
}

func (r *ResourcesServer) CreateResource(c *gin.Context) {
	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	// TODO: Implement create resource
	_ = authCtx

	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}

func (r *ResourcesServer) UpdateResource(c *gin.Context) {
	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	// TODO: Implement update resource
	_ = authCtx

	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
}
