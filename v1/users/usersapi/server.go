package usersapi

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ooqls/getset/email"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	v1 "github.com/ooqls/go-auth/v1"
	"github.com/ooqls/go-auth/v1/gen"
	"github.com/ooqls/go-auth/v1/users"
	"github.com/ooqls/go-auth/v1/users/usersapi/gen_users"
	"go.uber.org/zap"
)

var _ gen_users.ServerInterface = &UsersServer{}

type UsersServer struct {
	service     users.Service
	emailClient email.EmailClient
	l           *zap.Logger
}

func NewUsersServer(service users.Service, emailClient email.EmailClient, l *zap.Logger) *UsersServer {
	return &UsersServer{
		service:     service,
		emailClient: emailClient,
		l:           l,
	}
}

func (s *UsersServer) CreateUser(ctx *gin.Context) {
	var request gen_users.CreateUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}
	if request.Key != nil && request.Salt != nil {

		user, err := s.service.CreateUserWithPassword(authCtx, request.Email, request.Username, *request.Key, *request.Salt)
		if err != nil {
			v1.GinHandleError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, user)
		return
	} else {
		user, pw, err := s.service.CreateUserWithRandomPassword(authCtx, request.Email, request.Username)
		if err != nil {
			v1.GinHandleError(ctx, err)
			return
		}

		err = s.emailClient.SendHTMLEmail(ctx, email.EmailRecepient{
			Name:  request.Username,
			Email: request.Email,
		}, "New User", fmt.Sprintf("<b>Your password is %s</b>", pw))
		if err != nil {
			s.l.Error("failed to send new user email", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send email"})
			return
		}

		ctx.JSON(http.StatusOK, user)
		return
	}
}

func (s *UsersServer) GetUserByUsername(ctx *gin.Context, username string) {
	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	user, err := s.service.GetUserByUsername(authCtx, username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Silence unused variable warning
	_ = authCtx

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func (s *UsersServer) GetUserByID(ctx *gin.Context, id uuid.UUID) {
	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	user, err := s.service.GetUser(authCtx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Silence unused variable warning
	_ = authCtx

	ctx.JSON(http.StatusOK, gin.H{"user": user})
}

func (s *UsersServer) GetCurrentUser(ctx *gin.Context) {
	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	user, err := s.service.GetUser(authCtx, authCtx.GetAuthedUser().Id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	userResp := gen.User{
		Id:        user.Id,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	ctx.JSON(http.StatusOK, gin.H{"user": userResp})
}

func (s *UsersServer) UpdateUserByID(ctx *gin.Context, id uuid.UUID) {
	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	var request gen.UpdateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var emailVal string

	for _, ev := range request.Events {
		switch ev.Event {
		case "email":
			if v, ok := ev.Data["email"]; ok {
				emailVal = v.(string)
			}
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid field in update", "field": ev.Event})
			return
		}
	}

	user, err := s.service.GetUser(authCtx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	err = s.service.UpdateUserEmail(authCtx, id, emailVal)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Silence unused variable warning
	_ = authCtx

	ctx.JSON(http.StatusNoContent, nil)
}

func (s *UsersServer) DeleteUserByID(ctx *gin.Context, id uuid.UUID) {
	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}

	err := s.service.DeleteUser(authCtx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Silence unused variable warning
	_ = authCtx

	ctx.JSON(http.StatusOK, gin.H{"status": "user deleted"})
}

func (s *UsersServer) GetUsers(ctx *gin.Context, params gen_users.GetUsersParams) {
	authCtx, ok := authorizationv1.FromGinContext(ctx)
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Could not get authorization context"})
		return
	}
	page := 0
	pageSize := 10
	if params.Page != nil {
		page = *params.Page
	}
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	users, err := s.service.GetUsers(authCtx, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Silence unused variable warning
	_ = authCtx

	ctx.JSON(http.StatusOK, gin.H{"users": users})
}
