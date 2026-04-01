package rest

import (
	"app/domain/model/auth"
	"app/domain/model/response"
	"app/internal/rest/middleware"
	"app/matchmaking"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MatchmakingHandler struct {
	Service    *matchmaking.Service
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewMatchmakingHandler(route *gin.RouterGroup, svc *matchmaking.Service, mdl middleware.Middleware) {
	handler := &MatchmakingHandler{
		Service:    svc,
		Route:      route,
		Middleware: mdl,
	}

	mmRoute := route.Group("/matchmaking")
	mmRoute.Use(mdl.Auth())
	mmRoute.POST("/join", handler.JoinQueue)
	mmRoute.DELETE("/leave", handler.LeaveQueue)
	mmRoute.GET("/status", handler.QueueStatus)
}

func (h *MatchmakingHandler) JoinQueue(c *gin.Context) {
	claim := c.MustGet("token_data").(auth.JWTClaimUser)
	userID, _ := uuid.Parse(claim.UserID)

	// TODO: Fetch actual display name and ELO from user service/DB
	err := h.Service.JoinQueue(matchmaking.QueueEntry{
		UserID:      userID,
		DisplayName: "Player", // should come from DB
		EloRating:   1000,     // should come from DB
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(400, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"message":    "joined matchmaking queue",
		"queue_size": h.Service.QueueSize(),
	}))
}

func (h *MatchmakingHandler) LeaveQueue(c *gin.Context) {
	claim := c.MustGet("token_data").(auth.JWTClaimUser)
	userID, _ := uuid.Parse(claim.UserID)
	h.Service.LeaveQueue(userID)

	c.JSON(http.StatusOK, response.Success(map[string]string{
		"message": "left matchmaking queue",
	}))
}

func (h *MatchmakingHandler) QueueStatus(c *gin.Context) {
	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"queue_size": h.Service.QueueSize(),
	}))
}
