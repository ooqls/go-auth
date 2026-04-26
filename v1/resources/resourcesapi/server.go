package resourcesapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/resourcesv1"
	"github.com/ooqls/go-auth/v1/resources"
	"github.com/ooqls/go-auth/v1/resources/resourcesapi/gen_resources"
	"go.uber.org/zap"
)

var _ gen_resources.ServerInterface = &ResourcesServer{}

type ResourcesServer struct {
	service resources.Service
	l       *zap.Logger
}

func NewResourcesServer(factory datav1.Factory, l *zap.Logger) *ResourcesServer {
	return &ResourcesServer{
		service: resources.NewServiceImpl(factory),
		l:       l,
	}
}

func (r *ResourcesServer) DeleteResource(c *gin.Context, name string, group string, kind string) {
	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	err := r.service.DeleteResource(authCtx, group, kind, name)
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

	res, err := r.service.GetResource(authCtx, group, kind, name)
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

	page := int32(0)
	pageSize := int32(100)

	if params.Page != nil {
		page = int32(*params.Page)
	}
	if params.PageSize != nil {
		pageSize = int32(*params.PageSize)
	}

	res, err := r.service.GetResources(authCtx, params.Group, params.Kind, page, pageSize)
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
