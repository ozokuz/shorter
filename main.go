package main

import (
	"fmt"
	"log"
	"os"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ShortenedURL struct {
	gorm.Model
	Short string
	URL   string
}

func SetupRoutes(app *fiber.App) {
	// Index Page
	app.Get("/", func(c *fiber.Ctx) error {
		return RenderToHTML(c, IndexPage())
	})

	// Shorten URL / Success Page
	app.Post("/", func(c *fiber.Ctx) error {
		db := c.Locals("DB").(*gorm.DB)
		url := c.FormValue("url")

		short := Shorten()
		db.Create(&ShortenedURL{Short: short, URL: url})

		address := c.Locals("Address").(string)
		shortened := fmt.Sprintf("%s/s/%s", address, short)

		return RenderToHTML(c, SuccessPage(shortened))
	})

	// Redirect Shortened URL
	app.Get("/s/:short", func(c *fiber.Ctx) error {
		db := c.Locals("DB").(*gorm.DB)
		short := c.Params("short")

		var url ShortenedURL
		db.Where("short = ?", short).First(&url)

		if url.ID == 0 {
			return c.Status(fiber.StatusNotFound).SendString("Not found")
		}

		return c.Redirect(url.URL)
	})
}

func main() {
	// Connect to database
	db, err := gorm.Open(sqlite.Open("shortened.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&ShortenedURL{})

	// Create a new Fiber instance
	app := fiber.New()

	// Grab the host and port from the environment
	host := os.Getenv("HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Middleware to pass the database connection to the routes
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("DB", db)
		return c.Next()
	})

	// Middleware to set server address
	address := os.Getenv("SERVER_ADDRESS")
	if address == "" {
		address = fmt.Sprintf("http://%s:%s", host, port)
	}

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("Address", address)
		return c.Next()
	})

	// Setup routes
	SetupRoutes(app)

	// Serve static files
	app.Static("/static", "./public")

	// 404 Page
	app.Use(NotFoundMiddleware)

	// Start server
	listenAddr := fmt.Sprintf("%s:%s", host, port)
	log.Fatal(app.Listen(listenAddr))
}

func NotFoundMiddleware(c *fiber.Ctx) error {
	c.Status(fiber.StatusNotFound)
	return RenderToHTML(c, NotFound())
}

func RenderToHTML(c *fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html")
	shell := Shell(component)
	return shell.Render(c.Context(), c.Response().BodyWriter())
}
