package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"server/internal/usecase"
)

type StorageHandler struct {
	storageUC usecase.StorageUseCase
}

func NewStorageHandler(storageUC usecase.StorageUseCase) *StorageHandler {
	return &StorageHandler{storageUC: storageUC}
}

type ConnectYandexRequest struct {
	Code string `json:"code" binding:"required"`
}

func (h *StorageHandler) GetYandexAuthURL(c *gin.Context) {
	authURL := h.storageUC.GetAuthURL(c.Request.Context())
	
	c.JSON(http.StatusOK, gin.H{"auth_url": authURL})
}

func (h *StorageHandler) HandleYandexCallback(c *gin.Context) {
	userID := c.GetUint("userID")
	
	var req ConnectYandexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	err := h.storageUC.HandleCallback(c.Request.Context(), req.Code, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Yandex.Disk connected successfully"})
}

func (h *StorageHandler) GetFiles(c *gin.Context) {
	userID := c.GetUint("userID")
	path := c.DefaultQuery("path", "")
	
	files, err := h.storageUC.GetFiles(c.Request.Context(), userID, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"files": files})
}

func (h *StorageHandler) GetFileInfo(c *gin.Context) {
	userID := c.GetUint("userID")
	fileID := c.Param("id")
	
	// Конвертируем fileID в uint
	var id uint
	if _, err := fmt.Sscanf(fileID, "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}
	
	file, err := h.storageUC.GetFileInfo(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"file": file})
}

func (h *StorageHandler) DeleteFile(c *gin.Context) {
	userID := c.GetUint("userID")
	fileID := c.Param("id")
	
	// Конвертируем fileID в uint
	var id uint
	if _, err := fmt.Sscanf(fileID, "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}
	
	err := h.storageUC.DeleteFile(c.Request.Context(), userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

func (h *StorageHandler) GetYandexToken(c *gin.Context) {
	userID := c.GetUint("userID")

	token, err := h.storageUC.GetYandexToken(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем токен фронтенду для прямого доступа к Яндекс.Диску
	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
	})
}