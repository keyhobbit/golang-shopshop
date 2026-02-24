package web

import (
	"net/http"

	"shoop-golang/database"
	"shoop-golang/internal/models"

	"github.com/labstack/echo/v4"
)

func webData(c echo.Context) map[string]any {
	data := map[string]any{
		"IsLoggedIn": c.Get("is_logged_in"),
		"UserName":   c.Get("user_name"),
		"CartCount":  c.Get("cart_count"),
	}
	if fs, ok := c.Get("flash_success").(string); ok {
		data["FlashSuccess"] = fs
	}
	if fe, ok := c.Get("flash_error").(string); ok {
		data["FlashError"] = fe
	}

	var company models.CompanyInfo
	database.DB.First(&company)
	data["Company"] = company

	var categories []models.Category
	database.DB.Where("is_active = ?", true).Order("sort_order ASC").Find(&categories)
	data["NavCategories"] = categories

	return data
}

func Home(c echo.Context) error {
	data := webData(c)
	data["Title"] = "Trang chủ"

	var banners []models.Banner
	database.DB.Where("is_active = ?", true).Order("sort_order ASC").Find(&banners)
	data["Banners"] = banners

	var featured []models.Product
	database.DB.Preload("Images").Where("is_active = ? AND is_featured = ?", true, true).Limit(8).Find(&featured)
	data["FeaturedProducts"] = featured

	var latest []models.Product
	database.DB.Preload("Images").Where("is_active = ?", true).Order("created_at DESC").Limit(8).Find(&latest)
	data["LatestProducts"] = latest

	return c.Render(http.StatusOK, "web/home/index", data)
}
