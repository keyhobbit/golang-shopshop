package admin

import (
	"net/http"

	"shoop-golang/database"
	"shoop-golang/internal/models"
	"shoop-golang/pkg/session"

	"github.com/labstack/echo/v4"
)

func adminData(c echo.Context) map[string]any {
	sess := session.GetAdminSession(c)
	data := map[string]any{
		"AdminName": sess.Values["admin_name"],
	}
	flashes := session.GetFlash(c, sess, session.FlashSuccess)
	if len(flashes) > 0 {
		data["FlashSuccess"] = flashes[0]
	}
	errFlashes := session.GetFlash(c, sess, session.FlashError)
	if len(errFlashes) > 0 {
		data["FlashError"] = errFlashes[0]
	}
	return data
}

func Dashboard(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Dashboard"
	data["Active"] = "dashboard"

	var productCount, orderCount, userCount, categoryCount int64
	database.DB.Model(&models.Product{}).Count(&productCount)
	database.DB.Model(&models.Order{}).Count(&orderCount)
	database.DB.Model(&models.User{}).Count(&userCount)
	database.DB.Model(&models.Category{}).Count(&categoryCount)

	var recentOrders []models.Order
	database.DB.Preload("User").Order("created_at DESC").Limit(5).Find(&recentOrders)

	data["ProductCount"] = productCount
	data["OrderCount"] = orderCount
	data["UserCount"] = userCount
	data["CategoryCount"] = categoryCount
	data["RecentOrders"] = recentOrders

	return c.Render(http.StatusOK, "admin/dashboard/index", data)
}
