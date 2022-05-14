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
		var category string
		var imageFilename string

		if err := rows.Scan(&id, &name, &category, &imageFilename); err != nil {
			return fmt.Errorf("loadItems failed: %w", err)
		}

		i := Item{id: id, Name: name, Category: category, ImageFilename: imageFilename}
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

	// Insert into DB
	r, err := db.Exec("INSERT INTO items (name, category, image_filename) values (?, ?, ?)", name, category, imageFilename)
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
