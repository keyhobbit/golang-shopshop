package admin

import (
	"net/http"

	"shoop-golang/database"
	"shoop-golang/internal/models"
	"shoop-golang/pkg/session"

	"github.com/labstack/echo/v4"
)

func OrderList(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Đơn hàng"
	data["Active"] = "orders"

	status := c.QueryParam("status")
	query := database.DB.Preload("User").Preload("Items").Preload("Items.Product").Order("created_at DESC")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var orders []models.Order
	query.Find(&orders)
	data["Orders"] = orders
	data["FilterStatus"] = status

	return c.Render(http.StatusOK, "admin/orders/index", data)
}

func OrderDetail(c echo.Context) error {
	data := adminData(c)
	data["Title"] = "Chi tiết đơn hàng"
	data["Active"] = "orders"

	var order models.Order
	if err := database.DB.Preload("User").Preload("Items").Preload("Items.Product").First(&order, "id = ?", c.Param("id")).Error; err != nil {
		return c.Redirect(http.StatusFound, "/orders")
	}
	data["Order"] = order

	return c.Render(http.StatusOK, "admin/orders/detail", data)
}

func OrderUpdateStatus(c echo.Context) error {
	newStatus := c.FormValue("status")
	validStatuses := map[string]bool{
		"pending": true, "confirmed": true, "shipping": true, "delivered": true, "cancelled": true,
	}
	if !validStatuses[newStatus] {
		return c.Redirect(http.StatusFound, "/orders")
	}

	database.DB.Model(&models.Order{}).Where("id = ?", c.Param("id")).Update("status", newStatus)

	sess := session.GetAdminSession(c)
	session.SetFlash(c, sess, session.FlashSuccess, "Đã cập nhật trạng thái đơn hàng")
	return c.Redirect(http.StatusFound, "/orders/"+c.Param("id"))
}
