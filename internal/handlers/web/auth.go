package web

import (
	"net/http"

	"shoop-golang/database"
	"shoop-golang/internal/models"
	"shoop-golang/pkg/session"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func Register(c echo.Context) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")

	if name == "" || email == "" || password == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Vui lòng điền đầy đủ thông tin"})
	}

	var existing models.User
	if err := database.DB.Where("email = ?", email).First(&existing).Error; err == nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Email đã được sử dụng"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "message": "Lỗi hệ thống"})
	}

	user := models.User{
		Name:     name,
		Email:    email,
		Password: string(hash),
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "message": "Không thể tạo tài khoản"})
	}

	sess := session.GetWebSession(c)
	sess.Values["user_id"] = user.ID
	sess.Values["user_name"] = user.Name
	sess.Save(c.Request(), c.Response())

	return c.JSON(http.StatusOK, map[string]interface{}{"success": true, "redirect": c.FormValue("redirect")})
}

func Login(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	var user models.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Email hoặc mật khẩu không đúng"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Email hoặc mật khẩu không đúng"})
	}

	sess := session.GetWebSession(c)
	sess.Values["user_id"] = user.ID
	sess.Values["user_name"] = user.Name
	sess.Save(c.Request(), c.Response())

	return c.JSON(http.StatusOK, map[string]interface{}{"success": true, "redirect": c.FormValue("redirect")})
}

func Logout(c echo.Context) error {
	sess := session.GetWebSession(c)
	sess.Values = make(map[interface{}]interface{})
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusFound, "/")
}
