package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/pkg/logger"
	"github.com/alfanzaky/eraflazz/pkg/utils"
	"github.com/alfanzaky/eraflazz/pkg/xresponse"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userRepo    domain.UserRepository
	authService domain.AuthService
}

func (h *AuthHandler) generateUniqueUsername(email string) string {
	base := strings.Split(strings.ToLower(strings.TrimSpace(email)), "@")[0]
	base = strings.TrimSpace(base)
	if base == "" {
		base = "user"
	}

	username := base
	suffix := 1

	for {
		existing, _ := h.userRepo.GetByUsername(username)
		if existing == nil {
			return username
		}

		username = fmt.Sprintf("%s%d", base, suffix)
		suffix++
	}
}

func NewAuthHandler(userRepo domain.UserRepository, authService domain.AuthService) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, authService: authService}
}

type registerRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Company  string `json:"company"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		xresponse.BadRequest(c, "Invalid payload: "+err.Error())
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if !utils.ValidateEmail(req.Email) {
		xresponse.BadRequest(c, "Email tidak valid")
		return
	}

	if len(req.Password) < 8 {
		xresponse.BadRequest(c, "Password minimal 8 karakter")
		return
	}

	if existing, _ := h.userRepo.GetByEmail(req.Email); existing != nil {
		xresponse.Conflict(c, "Email sudah terdaftar")
		return
	}

	hashedPassword := utils.HashPassword(req.Password)
	username := h.generateUniqueUsername(req.Email)
	fullName := req.Name

	user := &domain.User{
		ID:           utils.GenerateUUID(),
		Username:     username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     &fullName,
		Level:        domain.LevelAgent,
		IsActive:     true,
		IsVerified:   true,
		AllowDebt:    false,
		Balance:      0,
		CreditLimit:  0,
	}

	if err := h.userRepo.Create(user); err != nil {
		logger.Error("Failed to register user", logger.ErrorField(err))
		xresponse.InternalServerError(c, "Gagal membuat akun")
		return
	}

	xresponse.Created(c, "Registrasi berhasil", gin.H{"user_id": user.ID})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		xresponse.BadRequest(c, "Invalid payload: "+err.Error())
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil || user == nil {
		xresponse.Unauthorized(c, "Email atau password salah")
		return
	}

	if !utils.VerifyPassword(req.Password, user.PasswordHash) {
		xresponse.Unauthorized(c, "Email atau password salah")
		return
	}

	token, err := h.authService.GenerateAccessToken(user)
	if err != nil {
		logger.Error("Failed to generate token", logger.ErrorField(err))
		xresponse.InternalServerError(c, "Gagal membuat token")
		return
	}

	c.SetCookie("session-token", token, 24*60*60, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil",
		"token":   token,
	})
}
