package admin

import (
	"net/http"
	"strconv"

	"shoop-golang/database"
	"shoop-golang/internal/models"
	"shoop-golang/pkg/session"
	"shoop-golang/pkg/utils"

	"github.com/labstack/echo/v4"
)

func CategoryList(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Danh mục"
	data["Active"] = "categories"

	var categories []models.Category
	database.DB.Order("sort_order ASC").Find(&categories)
	data["Categories"] = categories

	return c.Render(http.StatusOK, "admin/categories/index", data)
}

func CategoryCreate(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Thêm danh mục"
	data["Active"] = "categories"
	return c.Render(http.StatusOK, "admin/categories/form", data)
}

func CategoryStore(c echo.Context) error {
	sortOrder, _ := strconv.Atoi(c.FormValue("sort_order"))
	cat := models.Category{
		Name:        c.FormValue("name"),
		Slug:        utils.Slugify(c.FormValue("name")),
		Description: c.FormValue("description"),
		SortOrder:   sortOrder,
		IsActive:    c.FormValue("is_active") == "on",
	}

	if err := database.DB.Create(&cat).Error; err != nil {
		data := adminData(c)
		data["Title"] = "Thêm danh mục"
		data["Active"] = "categories"
		data["Error"] = "Không thể tạo danh mục: " + err.Error()
		data["Category"] = cat
		return c.Render(http.StatusOK, "admin/categories/form", data)
	}

	sess := session.GetAdminSession(c)
	session.SetFlash(c, sess, session.FlashSuccess, "Đã tạo danh mục thành công")
	return c.Redirect(http.StatusFound, "/categories")
}

func CategoryEdit(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Sửa danh mục"
	data["Active"] = "categories"

	var cat models.Category
	if err := database.DB.First(&cat, "id = ?", c.Param("id")).Error; err != nil {
		return c.Redirect(http.StatusFound, "/categories")
	}
	data["Category"] = cat
	data["IsEdit"] = true

	return c.Render(http.StatusOK, "admin/categories/form", data)
}

func CategoryUpdate(c echo.Context) error {
	var cat models.Category
	if err := database.DB.First(&cat, "id = ?", c.Param("id")).Error; err != nil {
		return c.Redirect(http.StatusFound, "/categories")
	}

	sortOrder, _ := strconv.Atoi(c.FormValue("sort_order"))
	cat.Name = c.FormValue("name")
	cat.Slug = utils.Slugify(c.FormValue("name"))
	cat.Description = c.FormValue("description")
	cat.SortOrder = sortOrder
	cat.IsActive = c.FormValue("is_active") == "on"

	database.DB.Save(&cat)

	sess := session.GetAdminSession(c)
	session.SetFlash(c, sess, session.FlashSuccess, "Đã cập nhật danh mục")
	return c.Redirect(http.StatusFound, "/categories")
}

func CategoryDelete(c echo.Context) error {
	database.DB.Where("id = ?", c.Param("id")).Delete(&models.Category{})
	sess := session.GetAdminSession(c)
	session.SetFlash(c, sess, session.FlashSuccess, "Đã xóa danh mục")
	return c.Redirect(http.StatusFound, "/categories")
}
