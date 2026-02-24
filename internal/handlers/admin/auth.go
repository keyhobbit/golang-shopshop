package admin

import (
	"net/http"

	"shoop-golang/database"
	"shoop-golang/internal/models"
	"shoop-golang/pkg/session"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func LoginPage(c echo.Context) error {
	return c.Render(http.StatusOK, "admin/login", map[string]any{
		"Title": "Admin Login",
	})
}

func Login(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	var admin models.AdminUser
	if err := database.DB.Where("email = ? AND is_active = ?", email, true).First(&admin).Error; err != nil {
		return c.Render(http.StatusOK, "admin/login", map[string]any{
			"Title": "Admin Login",
			"Error": "Email hoặc mật khẩu không đúng",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return c.Render(http.StatusOK, "admin/login", map[string]any{
			"Title": "Admin Login",
			"Error": "Email hoặc mật khẩu không đúng",
		})
	}

	sess := session.GetAdminSession(c)
	sess.Values["admin_id"] = admin.ID
	sess.Values["admin_name"] = admin.Name
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusFound, "/admin/dashboard")
}

func Logout(c echo.Context) error {
	sess := session.GetAdminSession(c)
	sess.Values = make(map[interface{}]interface{})
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusFound, "/admin/login")
}
