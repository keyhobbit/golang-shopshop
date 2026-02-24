package admin

import (
	"net/http"

	"shoop-golang/database"
	"shoop-golang/internal/models"

	"github.com/labstack/echo/v4"
)

func UserList(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Khách hàng"
	data["Active"] = "users"

	var users []models.User
	database.DB.Order("created_at DESC").Find(&users)
	data["Users"] = users

	return c.Render(http.StatusOK, "admin/users/index", data)
}

func UserDetail(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Chi tiết khách hàng"
	data["Active"] = "users"

	var user models.User
	if err := database.DB.Preload("Orders").First(&user, "id = ?", c.Param("id")).Error; err != nil {
		return c.Redirect(http.StatusFound, "/users")
	}
	data["User"] = user

	return c.Render(http.StatusOK, "admin/users/detail", data)
}
