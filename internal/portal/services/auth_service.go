package services

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/interfaces"
	"github.com/mikejsmith1985/devsmith-modular-platform/internal/portal/models"
	"github.com/rs/zerolog"
)

type AuthService struct {
	userRepo     interfaces.UserRepository
	githubClient interfaces.GitHubClient
	jwtSecret    []byte
	tokenExpiry  time.Duration
	logger       *zerolog.Logger
}

func NewAuthService(userRepo interfaces.UserRepository, githubClient interfaces.GitHubClient, jwtSecret string, logger *zerolog.Logger) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		githubClient: githubClient,
		jwtSecret:    []byte(jwtSecret),
		tokenExpiry:  24 * time.Hour,
		logger:       logger,
	}
}

func (s *AuthService) AuthenticateWithGitHub(ctx context.Context, code string) (*models.User, string, error) {
	token, err := s.githubClient.ExchangeCodeForToken(ctx, code)
	if err != nil {
		s.logger.Error().Err(err).Msg("GitHub token exchange failed")
		return nil, "", fmt.Errorf("token exchange failed: %w", err)
	}
	profile, err := s.githubClient.GetUserProfile(ctx, token)
	if err != nil {
		s.logger.Error().Err(err).Msg("GitHub profile fetch failed")
		return nil, "", fmt.Errorf("profile fetch failed: %w", err)
	}
	user := &models.User{
		GitHubID:          profile.ID,
		Username:          profile.Username,
		Email:             profile.Email,
		AvatarURL:         profile.AvatarURL,
		GitHubAccessToken: token,
	}
	err = s.userRepo.CreateOrUpdate(ctx, user)
	if err != nil {
		s.logger.Error().Err(err).Msg("User upsert failed")
		return nil, "", fmt.Errorf("user upsert failed: %w", err)
	}
	jwtToken, err := s.generateJWT(user)
	if err != nil {
		s.logger.Error().Err(err).Msg("JWT generation failed")
		return nil, "", fmt.Errorf("jwt generation failed: %w", err)
	}
	return user, jwtToken, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, token string) (*models.User, error) {
	claims := &jwt.RegisteredClaims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !tkn.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	userID := claims.Subject
	// Convert userID to int
	var id int
	_, err = fmt.Sscanf(userID, "%d", &id)
	if err != nil {
		return nil, fmt.Errorf("invalid user id in token: %w", err)
	}
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

func (s *AuthService) RevokeSession(ctx context.Context, token string) error {
	// For MVP, just let token expire (stateless JWT)
	// Future: implement session blacklist in DB
	return nil
}

func (s *AuthService) generateJWT(user *models.User) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", user.ID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenExpiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
