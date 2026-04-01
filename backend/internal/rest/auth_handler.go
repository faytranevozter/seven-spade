package rest

import (
	"app/domain/model/auth"
	"app/domain/model/response"
	"app/internal/rest/middleware"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	jwt_helper "app/helpers/jsonwebtoken"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

// OAuthUserInfo holds info extracted from OAuth providers
type OAuthUserInfo struct {
	Provider    string
	ProviderID  string
	Email       string
	DisplayName string
	AvatarURL   string
}

// UserUpsertFunc is called to find or create a user from OAuth info
type UserUpsertFunc func(ctx context.Context, info OAuthUserInfo) (userID uuid.UUID, displayName string, err error)

type AuthHandler struct {
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
	UpsertUser UserUpsertFunc

	googleOAuthConfig *oauth2.Config
	githubOAuthConfig *oauth2.Config
}

func NewAuthHandler(route *gin.RouterGroup, mdl middleware.Middleware, upsertUser UserUpsertFunc) *AuthHandler {
	handler := &AuthHandler{
		Route:      route,
		Middleware: mdl,
		UpsertUser: upsertUser,
	}

	// Google OAuth config
	handler.googleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("BACKEND_URL") + "/auth/google/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	// GitHub OAuth config
	handler.githubOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("BACKEND_URL") + "/auth/github/callback",
		Scopes:       []string{"read:user", "user:email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}

	authRoute := route.Group("/auth")
	authRoute.GET("/google", handler.GoogleLogin)
	authRoute.GET("/google/callback", handler.GoogleCallback)
	authRoute.GET("/github", handler.GithubLogin)
	authRoute.GET("/github/callback", handler.GithubCallback)
	authRoute.POST("/telegram", handler.TelegramLogin)
	authRoute.GET("/me", mdl.Auth(), handler.GetMe)

	return handler
}

// --- Google OAuth ---

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	state := uuid.NewString() // In production, store state in session/cookie to verify
	url := h.googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, response.Error(400, "missing code parameter"))
		return
	}

	token, err := h.googleOAuthConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		logrus.Errorf("Google OAuth exchange error: %v", err)
		c.JSON(http.StatusInternalServerError, response.Error(500, "failed to exchange token"))
		return
	}

	// Get user info from Google
	client := h.googleOAuthConfig.Client(c.Request.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		logrus.Errorf("Google userinfo error: %v", err)
		c.JSON(http.StatusInternalServerError, response.Error(500, "failed to get user info"))
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(500, "failed to decode user info"))
		return
	}

	h.completeOAuth(c, OAuthUserInfo{
		Provider:    "google",
		ProviderID:  googleUser.ID,
		Email:       googleUser.Email,
		DisplayName: googleUser.Name,
		AvatarURL:   googleUser.Picture,
	})
}

// --- GitHub OAuth ---

func (h *AuthHandler) GithubLogin(c *gin.Context) {
	state := uuid.NewString()
	url := h.githubOAuthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GithubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, response.Error(400, "missing code parameter"))
		return
	}

	token, err := h.githubOAuthConfig.Exchange(c.Request.Context(), code)
	if err != nil {
		logrus.Errorf("GitHub OAuth exchange error: %v", err)
		c.JSON(http.StatusInternalServerError, response.Error(500, "failed to exchange token"))
		return
	}

	// Get user info from GitHub
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(500, "failed to get user info"))
		return
	}
	defer resp.Body.Close()

	var ghUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(500, "failed to decode user info"))
		return
	}

	displayName := ghUser.Name
	if displayName == "" {
		displayName = ghUser.Login
	}

	h.completeOAuth(c, OAuthUserInfo{
		Provider:    "github",
		ProviderID:  fmt.Sprintf("%d", ghUser.ID),
		Email:       ghUser.Email,
		DisplayName: displayName,
		AvatarURL:   ghUser.AvatarURL,
	})
}

// --- Telegram Login ---

func (h *AuthHandler) TelegramLogin(c *gin.Context) {
	// Telegram sends login widget data as POST body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(400, "invalid request body"))
		return
	}

	var tgData struct {
		ID        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
		PhotoURL  string `json:"photo_url"`
		Hash      string `json:"hash"`
	}
	if err := json.Unmarshal(body, &tgData); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(400, "invalid telegram data"))
		return
	}

	// TODO: Verify Telegram hash using TELEGRAM_BOT_TOKEN
	// For now, trust the data (must implement HMAC verification in production)

	displayName := tgData.FirstName
	if tgData.LastName != "" {
		displayName += " " + tgData.LastName
	}

	h.completeOAuth(c, OAuthUserInfo{
		Provider:    "telegram",
		ProviderID:  fmt.Sprintf("%d", tgData.ID),
		Email:       "", // Telegram doesn't provide email
		DisplayName: displayName,
		AvatarURL:   tgData.PhotoURL,
	})
}

// --- Common ---

func (h *AuthHandler) completeOAuth(c *gin.Context, info OAuthUserInfo) {
	userID, displayName, err := h.UpsertUser(c.Request.Context(), info)
	if err != nil {
		logrus.Errorf("OAuth upsert error: %v", err)
		c.JSON(http.StatusInternalServerError, response.Error(500, "failed to create user"))
		return
	}

	// Generate JWT
	tokenString, err := jwt_helper.GenerateJWTToken(
		jwt_helper.GetJwtCredential().Member,
		auth.JWTClaimUser{
			UserID: userID.String(),
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(500, "failed to generate token"))
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	// Redirect to frontend with token
	c.Redirect(http.StatusTemporaryRedirect,
		fmt.Sprintf("%s/auth/callback?token=%s&name=%s", frontendURL, tokenString, displayName))
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	claim := c.MustGet("token_data").(auth.JWTClaimUser)
	c.JSON(http.StatusOK, response.Success(map[string]string{
		"user_id": claim.UserID,
	}))
}
