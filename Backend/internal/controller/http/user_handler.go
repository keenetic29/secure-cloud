package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"server/internal/usecase"
)

type UserHandler struct {
	userUC usecase.UserUseCase
}

func NewUserHandler(userUC usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUC: userUC,
	}
}

type UserProfileResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("userID")

	user, err := h.userUC.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	response := UserProfileResponse{
		ID:    user.ID,
		Email: user.Email,
	}

	c.JSON(http.StatusOK, response)
}

type UpdateProfileRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("userID")

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userUC.GetUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	user.Email = req.Email

	err = h.userUC.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}