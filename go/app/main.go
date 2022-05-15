package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir       = "image"
	jsonFileName = "items.json"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type Items struct {
	Items []Item `json:"items"`
}

var items Items

func loadItems() error {
	// Read items from JSON file
	f, err := os.Open(jsonFileName)
	if err != nil {
		return fmt.Errorf("loadItems failed: %w", err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&items); err != nil {
		return fmt.Errorf("loadItems failed: %w", err)
	}

	return nil
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	c.Logger().Infof("Receive item: %s, %s", name, category)

	// Add items
	items.Items = append(items.Items, Item{Name: name, Category: category})

	// Write to JSON file
	f, err := os.Create(jsonFileName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("addItem failed: %w", err))
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(items); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("addItem failed: %w", err))
	}

	// Response data
	message := fmt.Sprintf("item received: %s, %s", name, category)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("imageFilename"))

	if !strings.HasSuffix(imgPath, ".jpg") {
		res := Response{Message: "Image path does not end with .jpg"}
		return c.JSON(http.StatusBadRequest, res)
	}
	if _, err := os.Stat(imgPath); err != nil {
		c.Logger().Debugf("Image not found: %s", imgPath)
		imgPath = path.Join(ImgDir, "default.jpg")
	}
	return c.File(imgPath)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Logger.SetLevel(log.INFO)

	front_url := os.Getenv("FRONT_URL")
	if front_url == "" {
		front_url = "http://localhost:3000"
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{front_url},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
	}))

	// Load items from JSON file
	if err := loadItems(); err != nil {
		e.Logger.Fatal(err)
	}

	// Routes
	e.GET("/", root)
	e.POST("/items", addItem)
	e.GET("/image/:imageFilename", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
