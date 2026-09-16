package controllers

import (
	"net/http"

	"new-pos-api/models"

	"github.com/gin-gonic/gin"

	"new-pos-api/services"
	"new-pos-api/utils"
)

// Login godoc
// @Summary Login to get JWT token
// @Description Authenticates user and returns a JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body models.LoginRequest true "Login credentials"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/auth/login [post]
func Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid input")
		return
	}

	username := req.Username
	password := req.Password

	if username == "" || password == "" {
		utils.Error(c, http.StatusBadRequest, "Username and password required")
		return
	}

	auth, err := services.AuthenticateWithBranch(username, password)

	if err != nil {
		utils.Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	tokenData, err := utils.GetJWT(*auth, "user")
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.Success(c, "success", tokenData)
}
