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

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func BannerList(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Banner"
	data["Active"] = "banners"

	var banners []models.Banner
	database.DB.Order("sort_order ASC").Find(&banners)
	data["Banners"] = banners

	return c.Render(http.StatusOK, "admin/banners/index", data)
}

func BannerCreate(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Thêm Banner"
	data["Active"] = "banners"
	return c.Render(http.StatusOK, "admin/banners/form", data)
}

func BannerStore(c echo.Context) error {
	sortOrder, _ := strconv.Atoi(c.FormValue("sort_order"))
	banner := models.Banner{
		Title:     c.FormValue("title"),
		Subtitle:  c.FormValue("subtitle"),
		Link:      c.FormValue("link"),
		SortOrder: sortOrder,
		IsActive:  c.FormValue("is_active") == "on",
	}

	banner.Image = handleBannerUpload(c)

	if err := database.DB.Create(&banner).Error; err != nil {
		data := adminData(c)
		data["Title"] = "Thêm Banner"
		data["Active"] = "banners"
		data["Error"] = "Không thể tạo banner"
		return c.Render(http.StatusOK, "admin/banners/form", data)
	}

	sess := session.GetAdminSession(c)
	session.SetFlash(c, sess, session.FlashSuccess, "Đã tạo banner thành công")
	return c.Redirect(http.StatusFound, "/admin/banners")
}

func BannerEdit(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Sửa Banner"
	data["Active"] = "banners"

	var banner models.Banner
	if err := database.DB.First(&banner, "id = ?", c.Param("id")).Error; err != nil {
		return c.Redirect(http.StatusFound, "/admin/banners")
	}
	data["Banner"] = banner
	data["IsEdit"] = true

	return c.Render(http.StatusOK, "admin/banners/form", data)
}

func BannerUpdate(c echo.Context) error {
	var banner models.Banner
	if err := database.DB.First(&banner, "id = ?", c.Param("id")).Error; err != nil {
		return c.Redirect(http.StatusFound, "/admin/banners")
	}

	sortOrder, _ := strconv.Atoi(c.FormValue("sort_order"))
	banner.Title = c.FormValue("title")
	banner.Subtitle = c.FormValue("subtitle")
	banner.Link = c.FormValue("link")
	banner.SortOrder = sortOrder
	banner.IsActive = c.FormValue("is_active") == "on"

	if img := handleBannerUpload(c); img != "" {
		banner.Image = img
	}

	database.DB.Save(&banner)

	sess := session.GetAdminSession(c)
	session.SetFlash(c, sess, session.FlashSuccess, "Đã cập nhật banner")
	return c.Redirect(http.StatusFound, "/admin/banners")
}

func BannerDelete(c echo.Context) error {
	database.DB.Where("id = ?", c.Param("id")).Delete(&models.Banner{})
	sess := session.GetAdminSession(c)
	session.SetFlash(c, sess, session.FlashSuccess, "Đã xóa banner")
	return c.Redirect(http.StatusFound, "/admin/banners")
}

func handleBannerUpload(c echo.Context) string {
	file, err := c.FormFile("image")
	if err != nil {
		return ""
	}
	src, err := file.Open()
	if err != nil {
		return ""
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	uploadDir := "uploads/banners"
	os.MkdirAll(uploadDir, 0o755)
	dstPath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		return ""
	}
	defer dst.Close()
	io.Copy(dst, src)

	return "/" + dstPath
}
