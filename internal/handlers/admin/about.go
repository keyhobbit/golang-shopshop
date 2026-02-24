package admin

import (
	"net/http"

	"shoop-golang/database"
	"shoop-golang/internal/models"
	"shoop-golang/pkg/session"

	"github.com/labstack/echo/v4"
)

func AboutEdit(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Trang Giới thiệu"
	data["Active"] = "about"

	var about models.AboutPage
	database.DB.First(&about)
	data["About"] = about

	return c.Render(http.StatusOK, "admin/about/index", data)
}

func AboutUpdate(c echo.Context) error {
	var about models.AboutPage
	database.DB.First(&about)

	about.Title = c.FormValue("title")
	about.Content = c.FormValue("content")

	database.DB.Save(&about)

	sess := session.GetAdminSession(c)
	session.SetFlash(c, sess, session.FlashSuccess, "Đã cập nhật trang giới thiệu")
	return c.Redirect(http.StatusFound, "/admin/about")
}
