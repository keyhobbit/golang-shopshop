package web

import (
	"math"
	"net/http"
	"strconv"

	"shoop-golang/database"
	"shoop-golang/internal/models"

	"github.com/labstack/echo/v4"
)

func ProductList(c echo.Context) error {
	data := webData(c)
	data["Title"] = "Sản phẩm"

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage := 12
	offset := (page - 1) * perPage

	query := database.DB.Model(&models.Product{}).Where("is_active = ?", true)

	categorySlug := c.QueryParam("category")
	if categorySlug != "" {
		var cat models.Category
		if err := database.DB.Where("slug = ?", categorySlug).First(&cat).Error; err == nil {
			query = query.Where("category_id = ?", cat.ID)
			data["CurrentCategory"] = cat
		}
	}

	search := c.QueryParam("q")
	if search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
		data["SearchQuery"] = search
	}

	var total int64
	query.Count(&total)

	var products []models.Product
	query.Preload("Images").Preload("Category").Order("created_at DESC").Offset(offset).Limit(perPage).Find(&products)

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	data["Products"] = products
	data["Page"] = page
	data["TotalPages"] = totalPages
	data["Total"] = total

	return c.Render(http.StatusOK, "web/products/index", data)
}

func ProductDetail(c echo.Context) error {
	data := webData(c)

	var product models.Product
	if err := database.DB.Preload("Images").Preload("Category").Where("slug = ? AND is_active = ?", c.Param("slug"), true).First(&product).Error; err != nil {
		return c.Redirect(http.StatusFound, "/products")
	}

	data["Title"] = product.Name
	data["Product"] = product

	var related []models.Product
	database.DB.Preload("Images").Where("category_id = ? AND id != ? AND is_active = ?", product.CategoryID, product.ID, true).Limit(4).Find(&related)
	data["RelatedProducts"] = related

	return c.Render(http.StatusOK, "web/products/detail", data)
}
