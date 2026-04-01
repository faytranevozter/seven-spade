package rest

import (
	"app/domain/model/auth"
	request_model "app/domain/model/request"
	"app/domain/model/response"
	"app/internal/rest/middleware"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RoomService interface {
	CreateRoom(ctx context.Context, claim auth.JWTClaimUser, payload request_model.CreateRoomRequest) (int, response.Base)
	GetRoom(ctx context.Context, code string) (int, response.Base)
	JoinRoom(ctx context.Context, claim auth.JWTClaimUser, code string) (int, response.Base)
	KickPlayer(ctx context.Context, claim auth.JWTClaimUser, code string, targetUserID string) (int, response.Base)
	UpdateSettings(ctx context.Context, claim auth.JWTClaimUser, code string, payload request_model.UpdateRoomSettingsRequest) (int, response.Base)
}

type RoomHandler struct {
	Service    RoomService
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewRoomHandler(route *gin.RouterGroup, svc RoomService, mdl middleware.Middleware) {
	handler := &RoomHandler{
		Service:    svc,
		Route:      route,
		Middleware: mdl,
	}

	roomRoute := route.Group("/rooms")
	roomRoute.Use(mdl.Auth())
	roomRoute.POST("", handler.CreateRoom)
	roomRoute.GET("/:code", handler.GetRoom)
	roomRoute.POST("/:code/join", handler.JoinRoom)
	roomRoute.DELETE("/:code/players/:userId", handler.KickPlayer)
	roomRoute.PATCH("/:code/settings", handler.UpdateSettings)
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var payload request_model.CreateRoomRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid json data"))
		return
	}

	claim := c.MustGet("token_data").(auth.JWTClaimUser)
	statusCode, resp := h.Service.CreateRoom(c.Request.Context(), claim, payload)
	c.JSON(statusCode, resp)
}

func (h *RoomHandler) GetRoom(c *gin.Context) {
	statusCode, resp := h.Service.GetRoom(c.Request.Context(), c.Param("code"))
	c.JSON(statusCode, resp)
}

func (h *RoomHandler) JoinRoom(c *gin.Context) {
	claim := c.MustGet("token_data").(auth.JWTClaimUser)
	statusCode, resp := h.Service.JoinRoom(c.Request.Context(), claim, c.Param("code"))
	c.JSON(statusCode, resp)
}

func (h *RoomHandler) KickPlayer(c *gin.Context) {
	claim := c.MustGet("token_data").(auth.JWTClaimUser)
	statusCode, resp := h.Service.KickPlayer(c.Request.Context(), claim, c.Param("code"), c.Param("userId"))
	c.JSON(statusCode, resp)
}

func (h *RoomHandler) UpdateSettings(c *gin.Context) {
	var payload request_model.UpdateRoomSettingsRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(http.StatusBadRequest, "invalid json data"))
		return
	}

	claim := c.MustGet("token_data").(auth.JWTClaimUser)
	statusCode, resp := h.Service.UpdateSettings(c.Request.Context(), claim, c.Param("code"), payload)
	c.JSON(statusCode, resp)
}
