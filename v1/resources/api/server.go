package resourcesapi

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/datav1"
	"github.com/ooqls/go-auth/internal/resourcesv1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/ooqls/go-auth/v1/resources"
	"github.com/ooqls/go-auth/v1/resources/api/gen_resources"
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

	c.JSON(http.StatusOK, toGenResource(*res))
}

func (r *ResourcesServer) PostResourceSearch(c *gin.Context) {
	var req gen_resources.ResourceSearch
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	result, err := r.service.SearchResources(authCtx, req.Group, req.Kind, req.Name, req.Query)
	if err != nil {
		r.l.Error("failed to search resources", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, result)
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

	result, err := r.service.GetResources(authCtx, params.Group, params.Kind, page, pageSize)
	if err != nil {
		r.l.Error("failed to get resources", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gen_resources.ResourceList{
		Items:      toGenResourceList(result.Items),
		TotalCount: int(result.TotalCount),
	})
}

func (r *ResourcesServer) CreateResource(c *gin.Context) {
	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	var req gen_resources.CreateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid request body: %v", err)})
		return
	}
	res := req.Resource

	createdRes, err := r.service.CreateResource(authCtx, res.Group, res.Kind, res.Name)
	if err != nil {
		r.l.Error("failed to create resource", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, toGenResource(*createdRes))
}

func (r *ResourcesServer) UpdateResource(c *gin.Context) {
	authCtx, ok := authorizationv1.FromGinContext(c)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	// TODO: Implement update resource
	var req gen_resources.UpdateResource
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	res, err := r.service.UpdateResourceName(authCtx, req.Group, req.Kind, req.Name, req.NewName)
	if err != nil {
		if _, ok := err.(*v1.NotFoundError); ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
			return
		}
		r.l.Error("failed to update resource", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, toGenResource(*res))
}
