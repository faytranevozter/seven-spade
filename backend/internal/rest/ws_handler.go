package rest

import (
	"app/domain/model/auth"
	"app/domain/model/response"
	jwt_helper "app/helpers/jsonwebtoken"
	"app/internal/rest/middleware"
	"app/internal/ws"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		allowedOrigin := os.Getenv("WS_ORIGIN")
		if allowedOrigin == "" || allowedOrigin == "*" {
			return true
		}
		return r.Header.Get("Origin") == allowedOrigin
	},
}

type WSHandler struct {
	Hub        *ws.Hub
	Route      *gin.RouterGroup
	Middleware middleware.Middleware
}

func NewWSHandler(route *gin.RouterGroup, hub *ws.Hub, mdl middleware.Middleware) {
	handler := &WSHandler{
		Hub:        hub,
		Route:      route,
		Middleware: mdl,
	}

	wsRoute := route.Group("/ws")
	wsRoute.GET("/room/:code", handler.HandleRoomWS)
}

func (h *WSHandler) HandleRoomWS(c *gin.Context) {
	roomCode := c.Param("code")

	// Extract token from query param for WebSocket connections
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, response.Error(401, "token required"))
		return
	}

	// Validate JWT manually (since we can't use middleware for WS upgrade)
	claim, err := validateToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error(401, "invalid token"))
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Errorf("WebSocket upgrade error: %v", err)
		return
	}

	userID, _ := uuid.Parse(claim.UserID)
	displayName := c.Query("name")
	if displayName == "" {
		displayName = "Player"
	}

	client := ws.NewClient(conn, h.Hub, userID, displayName, roomCode)

	// Assign seat from query if reconnecting
	if seatStr := c.Query("seat"); seatStr != "" {
		var seat int
		if _, err := fmt.Sscanf(seatStr, "%d", &seat); err == nil {
			client.Seat = seat
		}
	}

	h.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

func validateToken(tokenString string) (*auth.JWTClaimUser, error) {
	secret := jwt_helper.GetJwtCredential().Member.Secret
	token, err := jwt.ParseWithClaims(tokenString, &auth.JWTClaimUser{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	claims, ok := token.Claims.(*auth.JWTClaimUser)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	return claims, nil
}
