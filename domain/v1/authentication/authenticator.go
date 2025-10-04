package authentication

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/go-auth/records"
	"github.com/ooqls/go-auth/records/v1/users"
	"github.com/ooqls/go-cache/cache"
	"github.com/ooqls/go-cache/factory"
	"github.com/ooqls/go-crypto/crypto"
	"github.com/ooqls/go-crypto/jwt"
	"github.com/ooqls/go-log"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var l *zap.Logger

func init() {
	l = log.NewLogger("authenticator")
}

type Registration struct {
	Username string
	Key      []byte
	Email    string
	Secret   []byte
}

var (
	ErrTokenExpired        error = errors.New("token expired")
	ErrInvalidToken        error = errors.New("token invalid")
	ErrInvalidPassword     error = errors.New("invalid password")
	ErrUserNotFound        error = errors.New("user not found")
	ErrUserExists          error = errors.New("user already exists")
	ErrInternal            error = errors.New("internal error")
	ErrRegistrationExpired error = errors.New("registration expired")
	ErrInvalidRegistration error = errors.New("invalid registration")
)

type Authenticator interface {
	ValidateRegistration(ctx context.Context, reg Registration) ([crypto.SALT_SIZE]byte, error)
	ChallengeRequest(ctx context.Context, user *users.User) (*Challenge, error)
	ChallengeResponse(ctx context.Context, challengeId uuid.UUID, solvedChallenge []byte) (okey string, rkey string, uid string, err error)
	AuthenticateNewUser(ctx context.Context, user *users.User) (*TokenResponse, error)
	AuthenticateWithToken(ctx context.Context, authToken string) (*TokenResponse, error)
	IsAuthenticated(ctx context.Context, token string) (*UserClaims, error)
}

type AuthenticatorV1 struct {
	tokenCache          cache.GenericCache
	authorizationIssuer jwt.TokenIssuer[UserClaims]
	refreshIssuer       jwt.TokenIssuer[UserClaims]
	challenger          Challenger
	audience            []string
}

func NewAuthenticatorV1(
	authorizationIssuer jwt.TokenIssuer[UserClaims],
	refreshIssuer jwt.TokenIssuer[UserClaims],
	cacheFactory factory.CacheFactory,
	challenger Challenger,
	audience []string) Authenticator {

	return &AuthenticatorV1{
		tokenCache:          cacheFactory.NewCache("token", 10*time.Minute),
		authorizationIssuer: authorizationIssuer,
		refreshIssuer:       refreshIssuer,
		challenger:          challenger,
		audience:            audience,
	}
}

func (a *AuthenticatorV1) ValidateRegistration(ctx context.Context, reg Registration) ([crypto.SALT_SIZE]byte, error) {
	l := l.With(zap.String("username", reg.Username))

	salt, err := a.challenger.VerifyRegistration(ctx, reg.Username, reg.Secret, reg.Key)
	if err != nil {
		l.Error("failed to verify registration", zap.Error(err))
		return [crypto.SALT_SIZE]byte{}, err
	}

	return salt, nil
}

// returns the auth token, refresh token and any errors
func (a *AuthenticatorV1) AuthenticateWithToken(ctx context.Context, authToken string) (*TokenResponse, error) {
	if authToken == "" || len(authToken) > 1024 {
		return nil, ErrInvalidToken
	}

	claims, err := a.getAuthTokenAuthentication(ctx, authToken)
	if err != nil {
		l.Error("failed to get user authentication", zap.Error(err))
		return nil, ErrInvalidToken
	}

	return &TokenResponse{
		UserId: claims.UserID,
	}, nil
}

// ChallengeRequest will issue a challenge for the given user
// Returns a challenge, and any errors
func (a *AuthenticatorV1) ChallengeRequest(ctx context.Context, user *users.User) (*Challenge, error) {
	challenge, err := a.challenger.IssueChallenge(ctx, user)
	if err != nil {
		return nil, err
	}

	return challenge, nil
}

// ChallengeResponse will verify the user's challenge response
// Returns the auth token, refresh token, user id, and any errors
func (a *AuthenticatorV1) ChallengeResponse(ctx context.Context, challengeId uuid.UUID, solvedChallenge []byte) (okey string, rkey string, uid string, err error) {
	result, err := a.challenger.VerifyChallenge(ctx, challengeId, solvedChallenge)
	if err != nil {
		return "", "", "", err
	}

	authToken, err := a.issueNewAuthToken(ctx, result.User.ID)
	if err != nil {
		return "", "", "", err
	}

	refreshToken, err := a.issueNewRefreshToken(ctx, result.User.ID)
	if err != nil {
		return "", "", "", err
	}

	return authToken, refreshToken, result.User.ID.String(), nil
}

// IsAuthenticated will check if a user is authenticated
// a user is authenticated when:
// 1. the token is valid
// 2. the token is not expired
// 3. the token is not revoked
// 4. the token is not blacklisted
func (a *AuthenticatorV1) IsAuthenticated(ctx context.Context, token string) (*UserClaims, error) {
	claims, err := a.getAuthTokenAuthentication(ctx, token)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func (a *AuthenticatorV1) AuthenticateNewUser(ctx context.Context, user *users.User) (*TokenResponse, error) {
	authToken, err := a.issueNewAuthToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := a.issueNewRefreshToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AuthToken:    authToken,
		RefreshToken: refreshToken,
		UserId:       user.ID,
	}, nil
}

func (a *AuthenticatorV1) issueNewAuthToken(ctx context.Context, id records.UserId) (string, error) {
	claims := UserClaims{UserID: id}
	token, _, err := a.authorizationIssuer.IssueToken(id.String(), UserClaims{UserID: id})
	if err != nil {
		return "", err
	}

	err = a.tokenCache.Set(ctx, token, claims)
	if err != nil {
		l.Error("failed to set token in cache", zap.Error(err))
	}

	return token, nil
}

func (a *AuthenticatorV1) issueNewRefreshToken(ctx context.Context, id records.UserId) (string, error) {
	claims := UserClaims{UserID: id}
	token, _, err := a.refreshIssuer.IssueToken(id.String(), claims)
	if err != nil {
		return "", err
	}

	err = a.tokenCache.Set(ctx, token, claims)
	if err != nil {
		l.Error("failed to set refresh token in cache", zap.Error(err))
	}

	return token, nil
}

func (a *AuthenticatorV1) getRefreshTokenAuthentication(ctx context.Context, token string) (*UserClaims, error) {
	var cachedClaims UserClaims
	err := a.tokenCache.Get(ctx, token, &cachedClaims)
	if err != nil && err != redis.Nil {
		l.Error("failed to get user authentication from cache", zap.Error(err))
	}

	if err == nil {
		return &cachedClaims, nil
	}
	jwtToken, claims, err := a.refreshIssuer.Decrypt(token)
	if err != nil {
		l.Error("failed to decrypt jwt token", zap.Error(err))
		return nil, ErrInvalidToken
	}

	if !jwtToken.Valid {
		l.Error("token is not valid")
		return nil, ErrInvalidToken
	}

	err = a.tokenCache.Set(ctx, token, claims)
	if err != nil {
		l.Error("failed to set user authentication in cache", zap.Error(err))
	}

	return &claims, nil
}

func (a *AuthenticatorV1) getAuthTokenAuthentication(ctx context.Context, token string) (*UserClaims, error) {
	var cachedClaims UserClaims
	err := a.tokenCache.Get(ctx, token, &cachedClaims)
	if err != nil && err != redis.Nil {
		l.Error("failed to get user authentication from cache", zap.Error(err))
	}

	if err == nil {
		return &cachedClaims, nil
	}

	jwtToken, claims, err := a.authorizationIssuer.Decrypt(token)
	if err != nil {
		l.Error("failed to decrypt jwt token", zap.Error(err))
		return nil, ErrInvalidToken
	}

	if !jwtToken.Valid {
		l.Error("token is not valid")
		return nil, ErrInvalidToken
	}

	err = a.tokenCache.Set(ctx, token, claims)
	if err != nil {
		l.Error("failed to set user authentication in cache", zap.Error(err))
	}

	return &claims, nil
}
