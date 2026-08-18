package authenticationv1

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/ooqls/getset/cache/cache"
	"github.com/ooqls/getset/cache/factory"
	"github.com/ooqls/getset/crypto/crypto"
	"github.com/ooqls/getset/log"
	"github.com/ooqls/go-auth/internal/authenticationv1/claims"
	"github.com/ooqls/go-auth/internal/contexts"
	"github.com/ooqls/go-auth/internal/usersv1"
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

type Authenticator interface {
	ValidateRegistration(ctx contexts.LContext, reg Registration) ([crypto.SALT_SIZE]byte, error)
	ChallengeRequest(ctx contexts.LContext, user *usersv1.User) (*Challenge, error)
	ChallengeResponse(ctx contexts.LContext, challengeId uuid.UUID, solvedChallenge []byte) (okey string, rkey string, uid string, err error)
	AuthenticateNewUser(ctx contexts.LContext, user *usersv1.User) (*TokenResponse, error)
	AuthenticateWithToken(ctx contexts.LContext, authToken string) (*TokenResponse, error)
	IsAuthenticated(ctx contexts.LContext, token string) (*claims.UserClaims, error)
}

type AuthenticatorV1 struct {
	tokenCache      cache.GenericCache
	userTokenIssuer *UserTokenIssuer
	challenger      Challenger
	audience        []string
}

func NewAuthenticatorV1(
	userTokenIssuer *UserTokenIssuer,
	cacheFactory factory.CacheFactory,
	challenger Challenger,
	audience []string) Authenticator {

	return &AuthenticatorV1{
		tokenCache:      cacheFactory.NewCache("token", 10*time.Minute),
		userTokenIssuer: userTokenIssuer,
		challenger:      challenger,
		audience:        audience,
	}
}

func (a *AuthenticatorV1) ValidateRegistration(ctx contexts.LContext, reg Registration) ([crypto.SALT_SIZE]byte, error) {
	l := l.With(zap.String("username", reg.Username))

	l.Info("validating registration")
	salt, err := a.challenger.VerifyRegistration(ctx, reg.Username, reg.Secret, reg.Key)
	if err != nil {
		l.Error("failed to verify registration", zap.Error(err))
		return [crypto.SALT_SIZE]byte{}, err
	}

	return salt, nil
}

// returns the auth token, refresh token and any errors
func (a *AuthenticatorV1) AuthenticateWithToken(ctx contexts.LContext, authToken string) (*TokenResponse, error) {
	ctx.L().Info("authenticating with token")
	if authToken == "" || len(authToken) > 2048 {
		return nil, ErrInvalidToken
	}

	ctx.L().Info("getting auth token claims")
	claims, err := a.getAuthTokenAuthentication(ctx, authToken)
	if err != nil {
		l.Error("failed to get user authentication", zap.Error(err))
		return nil, ErrInvalidToken
	}

	return &TokenResponse{
		UserId: uuid.UUID(claims.UserID),
	}, nil
}

// ChallengeRequest will issue a challenge for the given user
// Returns a challenge, and any errors
func (a *AuthenticatorV1) ChallengeRequest(ctx contexts.LContext, user *usersv1.User) (*Challenge, error) {
	ctx.L().Info("issuing challenge", zap.String("user_id", user.Id.String()))
	challenge, err := a.challenger.IssueChallenge(ctx, user)
	if err != nil {
		return nil, err
	}

	return challenge, nil
}

// ChallengeResponse will verify the user's challenge response
// Returns the auth token, refresh token, user id, and any errors
func (a *AuthenticatorV1) ChallengeResponse(ctx contexts.LContext, challengeId uuid.UUID, solvedChallenge []byte) (okey string, rkey string, uid string, err error) {
	ctx.L().Info("verifying challenge response", zap.String("challenge_id", challengeId.String()))
	result, err := a.challenger.VerifyChallenge(ctx, challengeId, solvedChallenge)
	if err != nil {
		return "", "", "", err
	}

	ctx.L().Info("challenge verified, issuing tokens", zap.String("challenge_id", challengeId.String()), zap.String("user_id", result.User.Id.String()))
	userData, err := json.Marshal(result.User)
	if err != nil {
		return "", "", "", err
	}

	ctx.L().Info("issuing auth token", zap.String("user_id", result.User.Id.String()))
	authToken, err := a.issueNewAuthToken(ctx, result.User.Id, userData)
	if err != nil {
		return "", "", "", err
	}

	ctx.L().Info("issuing refresh token", zap.String("user_id", result.User.Id.String()))
	refreshToken, err := a.issueNewRefreshToken(ctx, result.User.Id, userData)
	if err != nil {
		return "", "", "", err
	}

	return authToken, refreshToken, result.User.Object.Id.String(), nil
}

// IsAuthenticated will check if a user is authenticated
// a user is authenticated when:
// 1. the token is valid
// 2. the token is not expired
// 3. the token is not revoked
// 4. the token is not blacklisted
func (a *AuthenticatorV1) IsAuthenticated(ctx contexts.LContext, token string) (*claims.UserClaims, error) {
	claims, err := a.getAuthTokenAuthentication(ctx, token)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func (a *AuthenticatorV1) AuthenticateNewUser(ctx contexts.LContext, user *usersv1.User) (*TokenResponse, error) {
	userData, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	authToken, err := a.issueNewAuthToken(ctx, user.Id, userData)
	if err != nil {
		return nil, err
	}

	refreshToken, err := a.issueNewRefreshToken(ctx, user.Object.Id, userData)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AuthToken:    authToken,
		RefreshToken: refreshToken,
		UserId:       user.Object.Id,
	}, nil
}

func (a *AuthenticatorV1) issueNewAuthToken(ctx contexts.LContext, id uuid.UUID, userData []byte) (string, error) {
	claims := claims.UserClaims{
		UserID:   id,
		UserData: userData,
	}
	token, err := a.userTokenIssuer.IssueAuthorizationToken(ctx, claims)
	if err != nil {
		return "", err
	}

	err = a.tokenCache.Set(ctx, token, claims)
	if err != nil {
		l.Error("failed to set token in cache", zap.Error(err))
	}

	return token, nil
}

func (a *AuthenticatorV1) issueNewRefreshToken(ctx contexts.LContext, id uuid.UUID, userData []byte) (string, error) {
	claims := claims.UserClaims{
		UserID:   id,
		UserData: userData,
	}
	token, err := a.userTokenIssuer.IssueRefreshToken(ctx, claims)
	if err != nil {
		return "", err
	}

	err = a.tokenCache.Set(ctx, token, claims)
	if err != nil {
		l.Error("failed to set refresh token in cache", zap.Error(err))
	}

	return token, nil
}

func (a *AuthenticatorV1) getRefreshTokenAuthentication(ctx contexts.LContext, token string) (*claims.UserClaims, error) {
	var cachedClaims claims.UserClaims
	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := a.tokenCache.Get(cacheCtx, token, &cachedClaims)
	if err != nil && err != redis.Nil {
		l.Error("failed to get user authentication from cache", zap.Error(err))
	}

	if err == nil {
		return &cachedClaims, nil
	}

	jwtToken, claims, err := a.userTokenIssuer.refreshIssuer.Decrypt(token)
	if err != nil {
		l.Error("failed to decrypt jwt token", zap.Error(err))
		return nil, ErrInvalidToken
	}

	if !jwtToken.Valid {
		l.Error("token is not valid")
		return nil, ErrInvalidToken
	}

	setCtx, setCancel := context.WithTimeout(ctx, 5*time.Second)
	defer setCancel()
	err = a.tokenCache.Set(setCtx, token, claims)
	if err != nil {
		l.Error("failed to set user authentication in cache", zap.Error(err))
	}

	return &claims, nil
}

func (a *AuthenticatorV1) getAuthTokenAuthentication(ctx contexts.LContext, token string) (*claims.UserClaims, error) {
	var cachedClaims claims.UserClaims
	ctx.L().Debug("checking token cache for auth token", zap.String("token", token))
	cacheCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := a.tokenCache.Get(cacheCtx, token, &cachedClaims)
	if err != nil && !cache.IsCacheMissErr(err) {
		ctx.L().Error("failed to get user authentication from cache", zap.Error(err))
	}

	if err == nil {
		ctx.L().Debug("found auth token in cache", zap.String("token", token))
		return &cachedClaims, nil
	}

	jwtToken, claims, err := a.userTokenIssuer.DecryptAuthorizationToken(ctx, token)
	if err != nil {
		ctx.L().Error("failed to decrypt jwt token", zap.Error(err))
		return nil, ErrInvalidToken
	}

	if !jwtToken.Valid {
		ctx.L().Error("token is not valid")
		return nil, ErrInvalidToken
	}

	ctx.L().Debug("storing auth token in cache", zap.String("token", token))
	setCtx, setCancel := context.WithTimeout(ctx, 5*time.Second)
	defer setCancel()
	err = a.tokenCache.Set(setCtx, token, claims)
	if err != nil {
		ctx.L().Error("failed to set user authentication in cache", zap.Error(err))
	}

	return claims, nil
}
