package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"server/internal/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func Mount(r *gin.Engine, handler *Handler) {
	r.POST("/image", handler.SendImage)
}

func (h *Handler) SendImage(ctx *gin.Context) {
	file, fileHeader, err := ctx.Request.FormFile("image")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Couldn't get an image🚫",
		})
		return
	}

	id, err := h.service.LocalStack.UploadFile(file, fileHeader.Filename)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Couldn't connect with storage, pls try later💾",
		})
		return
	}

	err = h.service.SendMsg(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Couldn't connect 2nd service, pls try later💤",
		})
		return
	}

	msg, err := h.service.GetMessage()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Couldn't get msg from 2nd service, pls try later📛",
		})
		return
	}

	h.service.Logger.Infof(msg)

	var imgStruct service.Images

	err = json.Unmarshal([]byte(msg), &imgStruct)
	if err != nil {
		h.service.Logger.Infof("Couldn't parse responce:%s", err)
	}

	ctx.JSON(http.StatusOK, imgStruct)

}
