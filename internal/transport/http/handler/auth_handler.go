package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	secret      []byte
	adminUser   string
	adminPass   string
	tokenExpiry time.Duration
}

func NewAuthHandler(secret, adminUser, adminPass string, ttl time.Duration) *AuthHandler {
	return &AuthHandler{
		secret:      []byte(secret),
		adminUser:   adminUser,
		adminPass:   adminPass,
		tokenExpiry: ttl,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Username != h.adminUser || req.Password != h.adminPass {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	claims := jwt.MapClaims{
		"sub": req.Username,
		"exp": time.Now().Add(h.tokenExpiry).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
