package main

import (
	"log"

	"shoop-golang/config"
	"shoop-golang/database"
	"shoop-golang/database/seeders"
	adminHandlers "shoop-golang/internal/handlers/admin"
	"shoop-golang/internal/middleware"
	"shoop-golang/pkg/session"
	"shoop-golang/pkg/utils"

	"github.com/labstack/echo/v4"
	echoMw "github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.Load()
	db := database.Init(cfg.DBPath)
	seeders.Seed(db)
	session.Init(cfg.SessionSecret)

	e := echo.New()
	e.Renderer = utils.NewAdminRenderer("templates")

	e.Use(echoMw.Logger())
	e.Use(echoMw.Recover())
	e.Use(echoMw.GzipWithConfig(echoMw.GzipConfig{Level: 5}))

	e.Static("/static", "static")
	e.Static("/uploads", "uploads")

	e.GET("/admin/login", adminHandlers.LoginPage)
	e.POST("/admin/login", adminHandlers.Login)
	e.GET("/admin/logout", adminHandlers.Logout)

	admin := e.Group("/admin", middleware.AdminAuth)

	admin.GET("/dashboard", adminHandlers.Dashboard)
	admin.GET("", func(c echo.Context) error {
		return c.Redirect(301, "/admin/dashboard")
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

	log.Printf("Admin server starting on :%s", cfg.AdminPort)
	e.Logger.Fatal(e.Start(":" + cfg.AdminPort))
}
