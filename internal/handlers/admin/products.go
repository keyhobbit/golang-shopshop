package admin

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"shoop-golang/database"
	"shoop-golang/internal/models"
	"shoop-golang/pkg/session"
	"shoop-golang/pkg/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func ProductList(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Sản phẩm"
	data["Active"] = "products"

	var products []models.Product
	database.DB.Preload("Category").Preload("Images").Order("created_at DESC").Find(&products)
	data["Products"] = products

	return c.Render(http.StatusOK, "admin/products/index", data)
}

func ProductCreate(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Thêm sản phẩm"
	data["Active"] = "products"

	var categories []models.Category
	database.DB.Where("is_active = ?", true).Order("sort_order ASC").Find(&categories)
	data["Categories"] = categories

	return c.Render(http.StatusOK, "admin/products/form", data)
}

func ProductStore(c echo.Context) error {
	originalPrice, _ := strconv.ParseFloat(c.FormValue("original_price"), 64)
	salePrice, _ := strconv.ParseFloat(c.FormValue("sale_price"), 64)
	stock, _ := strconv.Atoi(c.FormValue("stock"))

	product := models.Product{
		Name:          c.FormValue("name"),
		Slug:          utils.Slugify(c.FormValue("name")),
		Description:   c.FormValue("description"),
		Content:       c.FormValue("content"),
		OriginalPrice: originalPrice,
		SalePrice:     salePrice,
		SKU:           c.FormValue("sku"),
		Stock:         stock,
		CategoryID:    c.FormValue("category_id"),
		IsActive:      c.FormValue("is_active") == "on",
		IsFeatured:    c.FormValue("is_featured") == "on",
	}

	if err := database.DB.Create(&product).Error; err != nil {
		data := adminData(c)
		data["Title"] = "Thêm sản phẩm"
		data["Active"] = "products"
		data["Error"] = "Không thể tạo sản phẩm: " + err.Error()
		var categories []models.Category
		database.DB.Where("is_active = ?", true).Find(&categories)
		data["Categories"] = categories
		data["Product"] = product
		return c.Render(http.StatusOK, "admin/products/form", data)
	}

	handleProductImages(c, product.ID)

	sess := session.GetAdminSession(c)
	session.SetFlash(c, sess, session.FlashSuccess, "Đã tạo sản phẩm thành công")
	return c.Redirect(http.StatusFound, "/products")
}

func ProductEdit(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Sửa sản phẩm"
	data["Active"] = "products"

	var product models.Product
	if err := database.DB.Preload("Images").First(&product, "id = ?", c.Param("id")).Error; err != nil {
		return c.Redirect(http.StatusFound, "/products")
	}
	data["Product"] = product
	data["IsEdit"] = true

	var categories []models.Category
	database.DB.Where("is_active = ?", true).Order("sort_order ASC").Find(&categories)
	data["Categories"] = categories

	return c.Render(http.StatusOK, "admin/products/form", data)
}

func ProductUpdate(c echo.Context) error {
	var product models.Product
	if err := database.DB.First(&product, "id = ?", c.Param("id")).Error; err != nil {
		return c.Redirect(http.StatusFound, "/products")
	}

	originalPrice, _ := strconv.ParseFloat(c.FormValue("original_price"), 64)
	salePrice, _ := strconv.ParseFloat(c.FormValue("sale_price"), 64)
	stock, _ := strconv.Atoi(c.FormValue("stock"))

	product.Name = c.FormValue("name")
	product.Slug = utils.Slugify(c.FormValue("name"))
	product.Description = c.FormValue("description")
	product.Content = c.FormValue("content")
	product.OriginalPrice = originalPrice
	product.SalePrice = salePrice
	product.SKU = c.FormValue("sku")
	product.Stock = stock
	product.CategoryID = c.FormValue("category_id")
	product.IsActive = c.FormValue("is_active") == "on"
	product.IsFeatured = c.FormValue("is_featured") == "on"

	database.DB.Save(&product)
	handleProductImages(c, product.ID)

	sess := session.GetAdminSession(c)
	session.SetFlash(c, sess, session.FlashSuccess, "Đã cập nhật sản phẩm")
	return c.Redirect(http.StatusFound, "/products")
}

func ProductDelete(c echo.Context) error {
	database.DB.Where("product_id = ?", c.Param("id")).Delete(&models.Image{})
	database.DB.Where("id = ?", c.Param("id")).Delete(&models.Product{})
	sess := session.GetAdminSession(c)
	session.SetFlash(c, sess, session.FlashSuccess, "Đã xóa sản phẩm")
	return c.Redirect(http.StatusFound, "/products")
}

func ImageDelete(c echo.Context) error {
	database.DB.Where("id = ?", c.Param("id")).Delete(&models.Image{})
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func handleProductImages(c echo.Context, productID string) {
	form, err := c.MultipartForm()
	if err != nil {
		return
	}
	files := form.File["images"]
	for i, file := range files {
		src, err := file.Open()
		if err != nil {
			continue
		}
		defer src.Close()

		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		uploadDir := "uploads/products"
		os.MkdirAll(uploadDir, 0o755)
		dstPath := filepath.Join(uploadDir, filename)

		dst, err := os.Create(dstPath)
		if err != nil {
			continue
		}
		defer dst.Close()
		io.Copy(dst, src)

		img := models.Image{
			ProductID: productID,
			URL:       "/" + dstPath,
			AltText:   file.Filename,
			SortOrder: i,
			IsPrimary: i == 0,
		}
		database.DB.Create(&img)
	}
}
