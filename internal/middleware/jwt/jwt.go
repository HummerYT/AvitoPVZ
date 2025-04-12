package jwt

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"AvitoPVZ/internal/models"
)

type Middleware struct {
	SecretKey string
}

func NewMiddleware(secretKey string) *Middleware {
	return &Middleware{
		SecretKey: secretKey,
	}
}

// SignedToken - подписание JWT для авторизированного пользователя токена
func (m *Middleware) SignedToken(ctx *fiber.Ctx) error {
	userID, ok := ctx.Context().Value("UserID").(uuid.UUID)
	if !ok {
		return ctx.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: "header user is empty",
		})
	}

	userStatus, ok := ctx.Context().Value("Role").(models.UserRole)
	if !ok {
		return ctx.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: "header user is empty",
		})
	}

	payload := jwt.MapClaims{
		"ExpiresAt": jwt.NewNumericDate(time.Now().UTC().Add(models.DurationJwtToken)),
		"IssuedAt":  jwt.NewNumericDate(time.Now().UTC()),
		"UserID":    userID,
		"Role":      userStatus,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	token.Header["kid"] = uuid.New().String()

	secretKey := []byte(m.SecretKey)

	jwtToken, err := token.SignedString(secretKey)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: err.Error(),
		})
	}

	ctx.Set(models.AuthorizationToken, jwtToken)

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"Token": jwtToken,
	})
}

func (m *Middleware) CompareToken(c *fiber.Ctx) error {
	tokenStr := c.Get(models.AuthorizationToken, "")
	if tokenStr == "" {
		return c.Status(http.StatusUnauthorized).JSON(models.ErrorResp{
			Message: "Token is empty",
		})
	}

	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	secretKey := []byte(m.SecretKey)

	jwtToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: fmt.Sprintf("JWT token is not valid: %v", err),
		})
	}

	payload, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok || !jwtToken.Valid {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: "Invalid token claims",
		})
	}

	expires, ok := payload["ExpiresAt"].(float64)
	if !ok {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: "Missing or invalid 'ExpiresAt' field",
		})
	}

	expiresAt := time.Unix(int64(expires), 0)
	if time.Now().After(expiresAt) {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: "Token is expired",
		})
	}

	userStatus, ok := payload["Role"].(string)
	if !ok {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: "Invalid status field",
		})
	}

	userID, ok := payload["UserID"].(string)
	if !ok {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: "Invalid status field",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{})
	}

	c.Locals("UserID", userUUID)
	c.Locals("Role", models.UserRole(userStatus))

	return c.Next()
}
