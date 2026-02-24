package testutil

import (
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"shoop-golang/database"
	"shoop-golang/database/seeders"
	adminHandlers "shoop-golang/internal/handlers/admin"
	webHandlers "shoop-golang/internal/handlers/web"
	"shoop-golang/internal/middleware"
	"shoop-golang/internal/models"
	"shoop-golang/pkg/session"
	"shoop-golang/pkg/utils"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	db.AutoMigrate(
		&models.AdminUser{},
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.Image{},
		&models.Order{},
		&models.OrderItem{},
		&models.Banner{},
		&models.CompanyInfo{},
		&models.AboutPage{},
		&models.SEOBanner{},
	)

	database.DB = db
	return db
}

func SetupTestDBWithSeed(t *testing.T) *gorm.DB {
	t.Helper()
	db := SetupTestDB(t)
	seeders.Seed(db)
	return db
}

func SetupSession() {
	session.Init("test-secret-key")
}

// NoopRenderer satisfies echo.Renderer for handler tests that call c.Render.
// It writes a minimal response so we can check status codes without real templates.
type NoopRenderer struct{}

func (n *NoopRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	w.Write([]byte("rendered:" + name))
	return nil
}

func NewAdminEcho() *echo.Echo {
	e := echo.New()
	e.Renderer = &NoopRenderer{}

	e.GET("/login", adminHandlers.LoginPage)
	e.POST("/login", adminHandlers.Login)
	e.GET("/logout", adminHandlers.Logout)

	admin := e.Group("", middleware.AdminAuth)
	admin.GET("/dashboard", adminHandlers.Dashboard)
	admin.GET("", func(c echo.Context) error {
		return c.Redirect(301, "/dashboard")
	})

	admin.GET("/categories", adminHandlers.CategoryList)
	admin.GET("/categories/create", adminHandlers.CategoryCreate)
	admin.POST("/categories", adminHandlers.CategoryStore)
	admin.GET("/categories/:id/edit", adminHandlers.CategoryEdit)
	admin.POST("/categories/:id", adminHandlers.CategoryUpdate)
	admin.POST("/categories/:id/delete", adminHandlers.CategoryDelete)

	admin.GET("/products", adminHandlers.ProductList)
	admin.GET("/products/create", adminHandlers.ProductCreate)
	admin.POST("/products", adminHandlers.ProductStore)
	admin.GET("/products/:id/edit", adminHandlers.ProductEdit)
	admin.POST("/products/:id", adminHandlers.ProductUpdate)
	admin.POST("/products/:id/delete", adminHandlers.ProductDelete)
	admin.POST("/images/:id/delete", adminHandlers.ImageDelete)

	admin.GET("/orders", adminHandlers.OrderList)
	admin.GET("/orders/:id", adminHandlers.OrderDetail)
	admin.POST("/orders/:id/status", adminHandlers.OrderUpdateStatus)

	admin.GET("/users", adminHandlers.UserList)
	admin.GET("/users/:id", adminHandlers.UserDetail)

	admin.GET("/banners", adminHandlers.BannerList)
	admin.GET("/banners/create", adminHandlers.BannerCreate)
	admin.POST("/banners", adminHandlers.BannerStore)
	admin.GET("/banners/:id/edit", adminHandlers.BannerEdit)
	admin.POST("/banners/:id", adminHandlers.BannerUpdate)
	admin.POST("/banners/:id/delete", adminHandlers.BannerDelete)

	admin.GET("/company", adminHandlers.CompanyEdit)
	admin.POST("/company", adminHandlers.CompanyUpdate)

	admin.GET("/about", adminHandlers.AboutEdit)
	admin.POST("/about", adminHandlers.AboutUpdate)

	return e
}

func NewWebEcho() *echo.Echo {
	e := echo.New()
	e.Renderer = &NoopRenderer{}
	e.Use(middleware.WebUserContext)

	e.GET("/", webHandlers.Home)
	e.GET("/products", webHandlers.ProductList)
	e.GET("/products/:slug", webHandlers.ProductDetail)
	e.POST("/register", webHandlers.Register)
	e.POST("/login", webHandlers.Login)
	e.GET("/logout", webHandlers.Logout)
	e.GET("/cart", webHandlers.CartPage)
	e.POST("/cart/add", webHandlers.AddToCart)
	e.POST("/cart/update", webHandlers.UpdateCart)
	e.POST("/checkout", webHandlers.Checkout)
	e.GET("/about", webHandlers.AboutPage)
	e.GET("/contact", webHandlers.ContactPage)

	return e
}

func NewWebRenderedEcho() *echo.Echo {
	e := echo.New()
	e.Renderer = utils.NewWebRenderer("../../templates")
	e.Use(middleware.WebUserContext)

	e.GET("/", webHandlers.Home)
	e.GET("/products", webHandlers.ProductList)
	e.GET("/products/:slug", webHandlers.ProductDetail)
	e.POST("/register", webHandlers.Register)
	e.POST("/login", webHandlers.Login)
	e.GET("/logout", webHandlers.Logout)
	e.GET("/cart", webHandlers.CartPage)
	e.POST("/cart/add", webHandlers.AddToCart)
	e.POST("/cart/update", webHandlers.UpdateCart)
	e.POST("/checkout", webHandlers.Checkout)
	e.GET("/about", webHandlers.AboutPage)
	e.GET("/contact", webHandlers.ContactPage)

	return e
}

func NewAdminRenderedEcho() *echo.Echo {
	e := echo.New()
	e.Renderer = utils.NewAdminRenderer("../../templates")

	e.GET("/login", adminHandlers.LoginPage)
	e.POST("/login", adminHandlers.Login)
	e.GET("/logout", adminHandlers.Logout)

	admin := e.Group("", middleware.AdminAuth)
	admin.GET("/dashboard", adminHandlers.Dashboard)

	admin.GET("/categories", adminHandlers.CategoryList)
	admin.POST("/categories", adminHandlers.CategoryStore)
	admin.GET("/categories/:id/edit", adminHandlers.CategoryEdit)
	admin.POST("/categories/:id", adminHandlers.CategoryUpdate)
	admin.POST("/categories/:id/delete", adminHandlers.CategoryDelete)

	admin.GET("/products", adminHandlers.ProductList)
	admin.POST("/products", adminHandlers.ProductStore)
	admin.GET("/products/:id/edit", adminHandlers.ProductEdit)
	admin.POST("/products/:id", adminHandlers.ProductUpdate)
	admin.POST("/products/:id/delete", adminHandlers.ProductDelete)

	admin.GET("/orders", adminHandlers.OrderList)
	admin.GET("/orders/:id", adminHandlers.OrderDetail)
	admin.POST("/orders/:id/status", adminHandlers.OrderUpdateStatus)

	admin.GET("/users", adminHandlers.UserList)
	admin.GET("/users/:id", adminHandlers.UserDetail)

	admin.GET("/banners", adminHandlers.BannerList)
	admin.POST("/banners", adminHandlers.BannerStore)
	admin.POST("/banners/:id", adminHandlers.BannerUpdate)
	admin.POST("/banners/:id/delete", adminHandlers.BannerDelete)

	admin.GET("/company", adminHandlers.CompanyEdit)
	admin.POST("/company", adminHandlers.CompanyUpdate)

	admin.GET("/about", adminHandlers.AboutEdit)
	admin.POST("/about", adminHandlers.AboutUpdate)

	return e
}

func CreateTestAdmin(t *testing.T) models.AdminUser {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := models.AdminUser{
		Email:    "admin@test.com",
		Password: string(hash),
		Name:     "Test Admin",
		Role:     "admin",
		IsActive: true,
	}
	database.DB.Create(&admin)
	return admin
}

func CreateTestUser(t *testing.T) models.User {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	user := models.User{
		Email:    "user@test.com",
		Password: string(hash),
		Name:     "Test User",
		Phone:    "0909111222",
		Address:  "123 Test St",
	}
	database.DB.Create(&user)
	return user
}

func CreateTestCategory(t *testing.T) models.Category {
	t.Helper()
	cat := models.Category{
		Name:     "Test Category",
		Slug:     "test-category",
		IsActive: true,
	}
	database.DB.Create(&cat)
	return cat
}

func CreateTestProduct(t *testing.T, categoryID string) models.Product {
	t.Helper()
	p := models.Product{
		Name:          "Test Product",
		Slug:          "test-product",
		Description:   "A test product",
		OriginalPrice: 100000,
		SalePrice:     80000,
		SKU:           "TEST-001",
		Stock:         10,
		CategoryID:    categoryID,
		IsActive:      true,
		IsFeatured:    true,
	}
	database.DB.Create(&p)
	return p
}

func CreateTestBanner(t *testing.T) models.Banner {
	t.Helper()
	b := models.Banner{
		Title:    "Test Banner",
		Subtitle: "Test subtitle",
		Image:    "/static/images/test.jpg",
		Link:     "/products",
		IsActive: true,
	}
	database.DB.Create(&b)
	return b
}

func CreateTestOrder(t *testing.T, userID, productID string) models.Order {
	t.Helper()
	order := models.Order{
		UserID:      userID,
		Status:      "pending",
		TotalAmount: 80000,
		Name:        "Test User",
		Phone:       "0909111222",
		Address:     "123 Test St",
		Items: []models.OrderItem{
			{ProductID: productID, Quantity: 1, Price: 80000},
		},
	}
	database.DB.Create(&order)
	return order
}

// AdminLoginCookies logs in as admin and returns cookies for authenticated requests.
func AdminLoginCookies(t *testing.T, ts *httptest.Server) []*http.Cookie {
	t.Helper()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}

	resp, err := client.PostForm(ts.URL+"/login", url.Values{
		"email":    {"admin@test.com"},
		"password": {"admin123"},
	})
	if err != nil {
		t.Fatalf("admin login failed: %v", err)
	}
	resp.Body.Close()
	return resp.Cookies()
}

// WebLoginCookies logs in as web user and returns cookies.
func WebLoginCookies(t *testing.T, ts *httptest.Server) []*http.Cookie {
	t.Helper()
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	resp, err := client.PostForm(ts.URL+"/login", url.Values{
		"email":    {"user@test.com"},
		"password": {"user123"},
	})
	if err != nil {
		t.Fatalf("web login failed: %v", err)
	}
	resp.Body.Close()
	return resp.Cookies()
}

func PostForm(ts *httptest.Server, path string, cookies []*http.Cookie, values url.Values) (*http.Response, error) {
	req, _ := http.NewRequest("POST", ts.URL+path, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	return client.Do(req)
}

func GetWithCookies(ts *httptest.Server, path string, cookies []*http.Cookie) (*http.Response, error) {
	req, _ := http.NewRequest("GET", ts.URL+path, nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}}
	return client.Do(req)
}
