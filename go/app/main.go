package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	ImgDir = "image"
	dbPath = "./db/mercari.sqlite3"
)

type Response struct {
	Message string `json:"message"`
}

type Item struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Category      string `json:"category"`
	ImageFilename string `json:"image_filename"`
}

var items []Item

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
	imagePath := c.FormValue("image")
	image, err := c.FormFile("image")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("addItem failed: %w", err))
	}

	// Hash the image file name
	h := sha256.Sum256([]byte(imagePath))
	imageHash := hex.EncodeToString(h[:])
	imageFilename := fmt.Sprintf("%s.jpg", imageHash)

	// Log
	c.Logger().Infof("Receive item: %s, %s, %s", name, category, imageFilename)

	// Open the image file
	imageFile, err := image.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("addItem failed: %w", err))
	}
	defer imageFile.Close()

	b := make([]byte, image.Size)
	if _, err := imageFile.Read(b); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("addItem failed: %w", err))
	}

	// Add the item to DB
	id, err := addItemToDB(name, category, imageFilename)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("addItem failed: %w", err))
	}

	// Update items
	i := Item{Id: id, Name: name, Category: category, ImageFilename: imageFilename}
	items = append(items, i)

	// Save the image to the local disk
	f, err := os.Create(fmt.Sprintf("%s/%d.jpg", ImgDir, id))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("addItem failed: %w", err))
	}
	defer f.Close()

	if _, err := f.Write(b); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("addItem failed: %w", err))
	}

	// Response data
	message := fmt.Sprintf("item received: %s, %s, %s", name, category, imageFilename)
	res := Response{Message: message}

	return c.JSON(http.StatusOK, res)
}

func getItem(c echo.Context) error {
	// Get item id
	id := c.Param("id")

	// Search for the specific item
	var item Item
	for _, i := range items {
		itemId, err := strconv.Atoi(id)
		if err != nil {
			return err
		}

		if i.Id == itemId {
			item = i
		}
	}

	return c.JSON(http.StatusOK, item)
}

func search(c echo.Context) error {
	k := c.QueryParam("keyword")
	var filteredItems []Item
	for _, i := range items {
		if strings.Contains(i.Name, k) {
			filteredItems = append(filteredItems, i)
		}
	}
	return c.JSON(http.StatusOK, map[string][]Item{"items": filteredItems})
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
	e.GET("/items/:id", getItem)
	e.GET("/search", search)
	e.GET("/image/:itemImg", getImg)

	// Start server
	e.Logger.Fatal(e.Start(":9000"))
}
