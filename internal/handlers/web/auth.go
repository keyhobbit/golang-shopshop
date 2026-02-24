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
	var req struct {
		Name     string `json:"name" form:"name"`
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
		Redirect string `json:"redirect" form:"redirect"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Dữ liệu không hợp lệ"})
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Vui lòng điền đầy đủ thông tin"})
	}

	var existing models.User
	if err := database.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Email đã được sử dụng"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "message": "Lỗi hệ thống"})
	}

	user := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hash),
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "message": "Không thể tạo tài khoản"})
	}

	sess := session.GetWebSession(c)
	sess.Values["user_id"] = user.ID
	sess.Values["user_name"] = user.Name
	sess.Save(c.Request(), c.Response())

	return c.JSON(http.StatusOK, map[string]interface{}{"success": true, "redirect": req.Redirect})
}

func Login(c echo.Context) error {
	var req struct {
		Email    string `json:"email" form:"email"`
		Password string `json:"password" form:"password"`
		Redirect string `json:"redirect" form:"redirect"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Dữ liệu không hợp lệ"})
	}

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Email hoặc mật khẩu không đúng"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "message": "Email hoặc mật khẩu không đúng"})
	}

	sess := session.GetWebSession(c)
	sess.Values["user_id"] = user.ID
	sess.Values["user_name"] = user.Name
	sess.Save(c.Request(), c.Response())

	return c.JSON(http.StatusOK, map[string]interface{}{"success": true, "redirect": req.Redirect})
}

func Logout(c echo.Context) error {
	sess := session.GetWebSession(c)
	sess.Values = make(map[interface{}]interface{})
	sess.Options.MaxAge = -1
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusFound, "/")
}
