package rest

import (
	"app/domain/model/auth"
	"app/domain/model/response"
	"app/helpers"
	gormrepo "app/internal/repository/gorm"
	"app/internal/rest/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GameHandler struct {
	GameRepo   *gormrepo.GameRepository
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewGameHandler(route *gin.RouterGroup, gameRepo *gormrepo.GameRepository, mdl middleware.Middleware) {
	handler := &GameHandler{
		GameRepo:   gameRepo,
		Route:      route,
		Middleware: mdl,
	}

	gameRoute := route.Group("/games")
	gameRoute.Use(mdl.Auth())
	gameRoute.GET("/history", handler.GameHistory)
	gameRoute.GET("/:id", handler.GameDetail)

	// Leaderboard (public)
	route.GET("/leaderboard", handler.Leaderboard)
}

func (h *GameHandler) GameHistory(c *gin.Context) {
	claim := c.MustGet("token_data").(auth.JWTClaimUser)
	page, limit, offset := helpers.GetLimitOffset(c.Request.URL.Query())

	games, total, err := h.GameRepo.FindGamesByUserID(c.Request.Context(), claim.UserID, int(limit), int(offset))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(500, err.Error()))
		return
	}

	list := make([]interface{}, len(games))
	for i, g := range games {
		list[i] = g
	}

	c.JSON(http.StatusOK, response.Success(response.List{
		List:  list,
		Page:  page,
		Limit: limit,
		Total: total,
	}))
}

func (h *GameHandler) GameDetail(c *gin.Context) {
	game, err := h.GameRepo.FindGameByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, response.Error(404, "game not found"))
		return
	}

	c.JSON(http.StatusOK, response.Success(game))
}

func (h *GameHandler) Leaderboard(c *gin.Context) {
	page, limit, offset := helpers.GetLimitOffset(c.Request.URL.Query())

	stats, total, err := h.GameRepo.GetLeaderboard(c.Request.Context(), int(limit), int(offset))
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(500, err.Error()))
		return
	}

	list := make([]interface{}, len(stats))
	for i, s := range stats {
		list[i] = s
	}

	c.JSON(http.StatusOK, response.Success(response.List{
		List:  list,
		Page:  page,
		Limit: limit,
		Total: total,
	}))
}
