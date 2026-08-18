package authentication

import (
	"errors"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/authenticationv1"
	"github.com/ooqls/go-auth/internal/authenticationv1/claims"
	"github.com/ooqls/go-auth/internal/authorizationv1"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/usersv1"
	"github.com/ooqls/go-auth/v1/users"
)

var (
	ErrInvalidUsername error = errors.New("invalid username")
)

type AuthedResult struct {
	Okey   string
	Rkey   string
	UserID string
}

type Service interface {
	Register(ctx contexts.LContext, reg authenticationv1.RegistrationParam) (*AuthedResult, error)
	AuthenticateNewUser(ctx contexts.LContext, user usersv1.User) (*authenticationv1.TokenResponse, error)
	AuthenticateWithToken(ctx contexts.LContext, authToken string) (*authenticationv1.TokenResponse, error)
	IssueChallenge(ctx contexts.LContext, username string) (*authenticationv1.Challenge, error)
	ValidateChallenge(ctx contexts.LContext, challengeId uuid.UUID, solvedChallenge []byte) (*AuthedResult, error)
	IsAuthenticated(ctx contexts.LContext, token string) (*claims.UserClaims, error)
}

type ServiceImpl struct {
	authenticator authenticationv1.Authenticator
	userSvc       users.Service
}

func NewServiceImpl(
	authenticator authenticationv1.Authenticator,
	userSvc users.Service) Service {
	return &ServiceImpl{
		authenticator: authenticator,
		userSvc:       userSvc,
	}
}

func (s *ServiceImpl) Register(ctx contexts.LContext, reg authenticationv1.RegistrationParam) (*AuthedResult, error) {
	salt, err := s.authenticator.ValidateRegistration(ctx, authenticationv1.Registration{
		Username: reg.Username,
		Key:      reg.Key,
		Email:    reg.Email,
		Secret:   reg.Secret,
	})
	if err != nil {
		return nil, err
	}
	authCtx := authorizationv1.NewInternalOperationContext(ctx)
	newUser, err := s.userSvc.CreateUserWithPassword(
		&authCtx,
		reg.Email,
		reg.Username,
		reg.Key,
		salt[:])
	if err != nil {
		return nil, err
	}
	authedResult, err := s.authenticator.AuthenticateNewUser(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return &AuthedResult{
		Okey:   authedResult.AuthToken,
		Rkey:   authedResult.RefreshToken,
		UserID: newUser.Id.String(),
	}, nil
}

func (s *ServiceImpl) AuthenticateNewUser(ctx contexts.LContext, user usersv1.User) (*authenticationv1.TokenResponse, error) {
	resp, err := s.authenticator.AuthenticateNewUser(ctx, &user)
	if err != nil {
		return nil, err
	}
	return &authenticationv1.TokenResponse{
		AuthToken:    resp.AuthToken,
		RefreshToken: resp.RefreshToken,
		UserId:       resp.UserId,
	}, nil
}

func (s *ServiceImpl) AuthenticateWithToken(ctx contexts.LContext, authToken string) (*authenticationv1.TokenResponse, error) {
	resp, err := s.authenticator.AuthenticateWithToken(ctx, authToken)
	if err != nil {
		return nil, err
	}
	return &authenticationv1.TokenResponse{
		AuthToken:    resp.AuthToken,
		RefreshToken: resp.RefreshToken,
		UserId:       resp.UserId,
	}, nil
}

func (s *ServiceImpl) IsAuthenticated(ctx contexts.LContext, token string) (*claims.UserClaims, error) {
	return s.authenticator.IsAuthenticated(ctx, token)
}

func (s *ServiceImpl) IssueChallenge(ctx contexts.LContext, username string) (*authenticationv1.Challenge, error) {
	authctx := authorizationv1.NewInternalOperationContext(ctx)
	user, err := s.userSvc.GetUserByUsername(&authctx, username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, ErrInvalidUsername
	}

	return s.authenticator.ChallengeRequest(ctx, user)
}

func (s *ServiceImpl) ValidateChallenge(ctx contexts.LContext, challengeId uuid.UUID, solvedChallenge []byte) (*AuthedResult, error) {
	okey, rkey, uid, err := s.authenticator.ChallengeResponse(ctx, challengeId, solvedChallenge)
	if err != nil {
		return nil, err
	}
	return &AuthedResult{
		Okey:   okey,
		Rkey:   rkey,
		UserID: uid,
	}, nil
}
