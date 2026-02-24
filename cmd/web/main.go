package main

import (
	"log"

	"shoop-golang/config"
	"shoop-golang/database"
	"shoop-golang/database/seeders"
	webHandlers "shoop-golang/internal/handlers/web"
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
	e.Renderer = utils.NewWebRenderer("templates")

	e.Use(echoMw.Logger())
	e.Use(echoMw.Recover())
	e.Use(echoMw.GzipWithConfig(echoMw.GzipConfig{Level: 5}))
	e.Use(middleware.WebUserContext)

	e.Static("/static", "static")
	e.Static("/uploads", "uploads")

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

	log.Printf("Web server starting on :%s", cfg.WebPort)
	e.Logger.Fatal(e.Start(":" + cfg.WebPort))
}
