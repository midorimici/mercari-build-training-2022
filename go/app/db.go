package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

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
		var categoryId int
		var imageFilename string

		if err := rows.Scan(&id, &name, &categoryId, &imageFilename); err != nil {
			return fmt.Errorf("loadItems failed: %w", err)
		}

		var category string
		categoryRow := db.QueryRow("SELECT name FROM category WHERE id = ?", categoryId)
		if err := categoryRow.Scan(&category); err != nil {
			return fmt.Errorf("loadItems failed: %w", err)
		}

		i := Item{Id: id, Name: name, Category: category, ImageFilename: imageFilename}
		items = append(items, i)
	}
	if err != nil {
		return fmt.Errorf("loadItems failed: %w", err)
	}

	return nil
}

func addItemToDB(name, category, imageFilename string) (int, error) {
	// Connect to DB
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return 0, fmt.Errorf("addItemToDB failed: %w", err)
	}
	defer db.Close()

	// Check if the category exists
	var categoryCount int
	countRow := db.QueryRow("SELECT COUNT(*) FROM category WHERE name = ?", category)
	if err := countRow.Scan(&categoryCount); err != nil {
		return 0, fmt.Errorf("addItemToDB failed: %w", err)
	}

	var categoryId int
	if categoryCount > 0 {
		// When the category already exists
		categoryRow := db.QueryRow("SELECT id FROM category WHERE name = ?", category)
		if err := categoryRow.Scan(&categoryId); err != nil {
			return 0, fmt.Errorf("addItemToDB failed: %w", err)
		}
	} else {
		// When the category does not exist
		// Create a new category in DB
		r, err := db.Exec("INSERT INTO category (name) values (?)", category)
		if err != nil {
			return 0, fmt.Errorf("addItemToDB failed: %w", err)
		}

		id, err := r.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("addItemToDB failed: %w", err)
		}

		categoryId = int(id)
	}

	// Insert into DB
	r, err := db.Exec("INSERT INTO items (name, category_id, image_filename) values (?, ?, ?)", name, categoryId, imageFilename)
	if err != nil {
		return 0, fmt.Errorf("addItemToDB failed: %w", err)
	}

	// Get last inserted item ID
	id, err := r.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("addItemToDB failed: %w", err)
	}

	return int(id), nil
}
