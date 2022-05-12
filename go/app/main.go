package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ImgDir = "image"
	dbPath = "./db/mercari.sqlite3"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	id       int
	Name     string `json:"name"`
	Category string `json:"category"`
}

var items []Item

func loadItems() error {
	// Read items from DB file
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("loadItems failed: %w", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM items")
	if err != nil {
		return fmt.Errorf("loadItems failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var name string
		var category string

		if err := rows.Scan(&id, &name, &category); err != nil {
			return fmt.Errorf("loadItems failed: %w", err)
		}

		i := Item{id: id, Name: name, Category: category}
		items = append(items, i)
	}
	if err != nil {
		return fmt.Errorf("loadItems failed: %w", err)
	}

	return nil
}

func root(c echo.Context) error {
	res := Response{Message: "Hello, world!"}
	return c.JSON(http.StatusOK, res)
}

func getItems(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string][]Item{"items": items})
}

func addItem(c echo.Context) error {
	// Get form data
	name := c.FormValue("name")
	category := c.FormValue("category")
	c.Logger().Infof("Receive item: %s, %s", name, category)

	// Write to DB
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("addItem failed: %w", err))
	}
	defer db.Close()

	r, err := db.Exec("INSERT INTO items (name, category) values (?, ?)", name, category)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("addItem failed: %w", err))
	}

	// Update items
	id, err := r.LastInsertId()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("addItem failed: %w", err))
	}
	i := Item{id: int(id), Name: name, Category: category}
	items = append(items, i)

	// Response data
	message := fmt.Sprintf("item received: %s, %s", name, category)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func getImg(c echo.Context) error {
	// Create image path
	imgPath := path.Join(ImgDir, c.Param("itemImg"))

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
	e.GET("/items", getItems)
	e.POST("/items", addItem)
	e.GET("/image/:itemImg", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
