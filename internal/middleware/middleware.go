package middleware

import (
	"encoding/json"
	"net/http"

	"shoop-golang/internal/models"
	"shoop-golang/pkg/session"

	"github.com/labstack/echo/v4"
)

func AdminAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess := session.GetAdminSession(c)
		adminID, ok := sess.Values["admin_id"].(string)
		if !ok || adminID == "" {
			return c.Redirect(http.StatusFound, "/login")
		}
		c.Set("admin_id", adminID)
		c.Set("admin_name", sess.Values["admin_name"])
		return next(c)
	}
}

func WebAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess := session.GetWebSession(c)
		userID, ok := sess.Values["user_id"].(string)
		if !ok || userID == "" {
			return c.Redirect(http.StatusFound, "/login")
		}
		c.Set("user_id", userID)
		c.Set("user_name", sess.Values["user_name"])
		return next(c)
	}
}

// Injects user info into context if logged in (non-blocking)
func WebUserContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess := session.GetWebSession(c)
		if userID, ok := sess.Values["user_id"].(string); ok && userID != "" {
			c.Set("user_id", userID)
			c.Set("user_name", sess.Values["user_name"])
			c.Set("is_logged_in", true)
		} else {
			c.Set("is_logged_in", false)
		}

		flashes := session.GetFlash(c, sess, session.FlashSuccess)
		if len(flashes) > 0 {
			c.Set("flash_success", flashes[0])
		}
		errFlashes := session.GetFlash(c, sess, session.FlashError)
		if len(errFlashes) > 0 {
			c.Set("flash_error", errFlashes[0])
		}

		// Cart count
		if data, ok := sess.Values["cart"].(string); ok && data != "" {
			var items []models.CartItem
			if json.Unmarshal([]byte(data), &items) == nil {
				count := 0
				for _, item := range items {
					count += item.Quantity
				}
				c.Set("cart_count", count)
			}
		}
		if c.Get("cart_count") == nil {
			c.Set("cart_count", 0)
		}

		return next(c)
	}
}
