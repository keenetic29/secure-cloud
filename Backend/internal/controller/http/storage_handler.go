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

type DownloadFileRequest struct {
	MasterPassword string `json:"master_password" binding:"required"`
}

type GetFilenameRequest struct {
	MasterPassword string `json:"master_password" binding:"required"`
}

type UploadFileResponse struct {
	FileID    uint   `json:"file_id"`
	Message   string `json:"message"`
	EncryptedName string `json:"encrypted_name"`
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
	path := c.DefaultQuery("path", "/")
	
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

func (h *StorageHandler) GetDecryptedFilename(c *gin.Context) {
	userID := c.GetUint("userID")
	fileID := c.Param("id")
	
	var id uint
	if _, err := fmt.Sscanf(fileID, "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}
	
	var req GetFilenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	filename, err := h.storageUC.GetDecryptedFilename(c.Request.Context(), userID, id, req.MasterPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"filename": filename})
}

func (h *StorageHandler) UploadFile(c *gin.Context) {
	userID := c.GetUint("userID")
	path := c.DefaultQuery("path", "/")
	masterPassword := c.PostForm("master_password")
	
	if masterPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "master password is required"})
		return
	}
	
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	
	// Проверяем размер файла (макс 100MB)
	if file.Size > 100*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file size exceeds 100MB limit"})
		return
	}
	
	metadata, err := h.storageUC.UploadFile(c.Request.Context(), userID, file, masterPassword, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	response := UploadFileResponse{
		FileID:    metadata.ID,
		Message:   "File uploaded and encrypted successfully",
		EncryptedName: metadata.EncryptedName,
	}
	
	c.JSON(http.StatusOK, response)
}

func (h *StorageHandler) DownloadFile(c *gin.Context) {
	userID := c.GetUint("userID")
	fileID := c.Param("id")
	
	var id uint
	if _, err := fmt.Sscanf(fileID, "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file ID"})
		return
	}
	
	var req DownloadFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	content, filename, err := h.storageUC.DownloadFile(c.Request.Context(), userID, id, req.MasterPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// Устанавливаем заголовки для скачивания
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", len(content)))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate")
	c.Header("Pragma", "public")
	
	c.Data(http.StatusOK, "application/octet-stream", content)
}

func (h *StorageHandler) DeleteFile(c *gin.Context) {
	userID := c.GetUint("userID")
	fileID := c.Param("id")
	
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

	c.JSON(http.StatusOK, gin.H{"access_token": token})
}