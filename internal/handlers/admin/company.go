package admin

import (
	"net/http"

	"shoop-golang/database"
	"shoop-golang/internal/models"
	"shoop-golang/pkg/session"

	"github.com/labstack/echo/v4"
)

func CompanyEdit(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Thông tin công ty"
	data["Active"] = "company"

	var info models.CompanyInfo
	database.DB.First(&info)
	data["Company"] = info

	return c.Render(http.StatusOK, "admin/company/index", data)
}

func CompanyUpdate(c echo.Context) error {
	var info models.CompanyInfo
	database.DB.First(&info)

	info.Name = c.FormValue("name")
	info.Tagline = c.FormValue("tagline")
	info.Email = c.FormValue("email")
	info.Phone = c.FormValue("phone")
	info.Address = c.FormValue("address")
	info.FacebookURL = c.FormValue("facebook_url")
	info.ZaloURL = c.FormValue("zalo_url")
	info.Copyright = c.FormValue("copyright")

	database.DB.Save(&info)

	sess := session.GetAdminSession(c)
	session.SetFlash(c, sess, session.FlashSuccess, "Đã cập nhật thông tin công ty")
	return c.Redirect(http.StatusFound, "/admin/company")
}
