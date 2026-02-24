package web

import (
	"net/http"

	"shoop-golang/database"
	"shoop-golang/internal/models"

	"github.com/labstack/echo/v4"
)

func AboutPage(c echo.Context) error {
	data := webData(c)
	data["Title"] = "Giới thiệu"

	var about models.AboutPage
	database.DB.First(&about)
	data["About"] = about

	return c.Render(http.StatusOK, "web/about/index", data)
}

func ContactPage(c echo.Context) error {
	data := webData(c)
	data["Title"] = "Liên hệ"
	return c.Render(http.StatusOK, "web/contact/index", data)
}
