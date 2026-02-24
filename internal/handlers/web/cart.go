package web

import (
	"encoding/json"
	"net/http"
	"strconv"

	"shoop-golang/database"
	"shoop-golang/internal/models"
	"shoop-golang/pkg/session"

	"github.com/labstack/echo/v4"
)

func CartPage(c echo.Context) error {
	data := webData(c)
	data["Title"] = "Giỏ hàng"

	items := getCartItems(c)
	data["CartItems"] = items

	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}
	data["CartTotal"] = total

	return c.Render(http.StatusOK, "web/cart/index", data)
}

func AddToCart(c echo.Context) error {
	isLoggedIn, _ := c.Get("is_logged_in").(bool)
	if !isLoggedIn {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "login_required"})
	}

	productID := c.FormValue("product_id")
	qty := 1
	if n, err := strconv.Atoi(c.FormValue("quantity")); err == nil && n > 0 {
		qty = n
	}
	var product models.Product
	if err := database.DB.Preload("Images").First(&product, "id = ?", productID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Sản phẩm không tồn tại"})
	}

	price := product.SalePrice
	if price <= 0 {
		price = product.OriginalPrice
	}

	imageURL := "/static/images/placeholder.jpg"
	for _, img := range product.Images {
		if img.IsPrimary {
			imageURL = img.URL
			break
		}
	}
	if imageURL == "/static/images/placeholder.jpg" && len(product.Images) > 0 {
		imageURL = product.Images[0].URL
	}

	items := getCartItems(c)

	found := false
	for i, item := range items {
		if item.ProductID == productID {
			items[i].Quantity += qty
			found = true
			break
		}
	}
	if !found {
		items = append(items, models.CartItem{
			ProductID: productID,
			Name:      product.Name,
			Image:     imageURL,
			Price:     price,
			Quantity:  qty,
		})
	}

	saveCartItems(c, items)

	return c.JSON(http.StatusOK, map[string]any{
		"status":    "ok",
		"cartCount": cartCount(items),
	})
}

func UpdateCart(c echo.Context) error {
	productID := c.FormValue("product_id")
	action := c.FormValue("action")

	items := getCartItems(c)
	for i, item := range items {
		if item.ProductID == productID {
			switch action {
			case "increase":
				items[i].Quantity++
			case "decrease":
				items[i].Quantity--
				if items[i].Quantity <= 0 {
					items = append(items[:i], items[i+1:]...)
				}
			case "remove":
				items = append(items[:i], items[i+1:]...)
			}
			break
		}
	}

	saveCartItems(c, items)

	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status":    "ok",
		"cartCount": cartCount(items),
		"cartTotal": total,
	})
}

func Checkout(c echo.Context) error {
	isLoggedIn, _ := c.Get("is_logged_in").(bool)
	if !isLoggedIn {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "login_required"})
	}

	items := getCartItems(c)
	if len(items) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Giỏ hàng trống"})
	}

	userID, _ := c.Get("user_id").(string)

	var total float64
	var orderItems []models.OrderItem
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
		orderItems = append(orderItems, models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		})
	}

	// Support both JSON and form
	var name, phone, address, note string
	if c.Request().Header.Get("Content-Type") == "application/json" {
		var body struct {
			Name    string `json:"name"`
			Phone   string `json:"phone"`
			Address string `json:"address"`
			Note    string `json:"note"`
		}
		if err := c.Bind(&body); err == nil {
			name, phone, address, note = body.Name, body.Phone, body.Address, body.Note
		}
	}
	if name == "" {
		name = c.FormValue("name")
	}
	if phone == "" {
		phone = c.FormValue("phone")
	}
	if address == "" {
		address = c.FormValue("address")
	}
	if note == "" {
		note = c.FormValue("note")
	}

	order := models.Order{
		UserID:      userID,
		Status:      "pending",
		TotalAmount: total,
		Name:        name,
		Phone:       phone,
		Address:     address,
		Note:        note,
		Items:       orderItems,
	}

	if err := database.DB.Create(&order).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Không thể tạo đơn hàng"})
	}

	saveCartItems(c, []models.CartItem{})

	sess := session.GetWebSession(c)
	session.SetFlash(c, sess, session.FlashSuccess, "Đặt hàng thành công! Mã đơn: "+order.ID[:8])

	return c.JSON(http.StatusOK, map[string]any{
		"success":  true,
		"order_id": order.ID,
		"redirect": "/",
		"message":  "Đặt hàng thành công! Mã đơn: " + order.ID[:8],
	})
}

func getCartItems(c echo.Context) []models.CartItem {
	sess := session.GetWebSession(c)
	data, ok := sess.Values["cart"].(string)
	if !ok {
		return []models.CartItem{}
	}
	var items []models.CartItem
	json.Unmarshal([]byte(data), &items)
	return items
}

func saveCartItems(c echo.Context, items []models.CartItem) {
	sess := session.GetWebSession(c)
	data, _ := json.Marshal(items)
	sess.Values["cart"] = string(data)
	sess.Save(c.Request(), c.Response())
}

func cartCount(items []models.CartItem) int {
	count := 0
	for _, item := range items {
		count += item.Quantity
	}
	return count
}
