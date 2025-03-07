package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var dbConn *pgx.Conn

func connect() (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), "postgres://tarunrawat:Tarun%40124@localhost:5432/mydb")
	if err != nil {
		return nil, err
	}
	fmt.Println("✅ Connected to PostgreSQL!")
	return conn, nil
}

func main() {
	greeting()
	var err error
	dbConn, err = pgx.Connect(context.Background(), "postgres://tarunrawat:Tarun%40124@localhost:5432/mydb")
	if err != nil {
		log.Fatalf("Database Connection Error: %v", err)
	}
	defer dbConn.Close(context.Background())
	fmt.Println("Connected to PostgreSQL!")

	//Initialize Gin Router
	router := gin.Default()
	router.POST("/shorten", shortenHandler)
	router.GET("/:shortcode", redirectHandler)

	//Start Server
	fmt.Println("Server running at http://localhost:8080")
	router.Run(":8080")

}

func greeting() {
	fmt.Println("Hello, World! Welcome to URL Shortener!")
}

func receieveUserInput() string {
	var userURL string
	fmt.Println("Enter a URL: ")
	fmt.Scan(&userURL)
	return userURL
}

func generateShortCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	shortCodeLength := 6
	shortCode := make([]byte, shortCodeLength)

	for i := range shortCode {
		shortCode[i] = charset[r.Intn(len(charset))]
	}

	return string(shortCode)

}

// Store Data in PostGreSQL
func storeURL(conn *pgx.Conn, longURL, shortCode string) error {
	query := `INSERT INTO urls (short_code, long_url) VALUES ($1, $2)`
	_, err := conn.Exec(context.Background(), query, shortCode, longURL)
	if err != nil {
		return err
	}
	fmt.Println("✅ URL stored successfully in the database.")
	return nil
}

// Retrieve the original long URL based on the shortcode
func getLongURL(conn *pgx.Conn, shortCode string) (string, error) {
	var longURL string
	query := "SELECT long_url FROM urls WHERE short_code = $1"
	err := conn.QueryRow(context.Background(), query, shortCode).Scan(&longURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			fmt.Printf("Shortcode %s not found in DB.\n", shortCode)
			return "", fmt.Errorf("shortcode not found")
		}
		fmt.Printf("Database query error: %v\n", err)
		return "", err
	}

	fmt.Printf("Shortcode %s maps to URL: %s\n", shortCode, longURL)
	return longURL, nil

}

// POST Request to shorten the incoming long_url
func shortenHandler(c *gin.Context) {
	var json struct {
		LongURL string `json:"long_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Payload"})
		return
	}

	// Generate Shortcode
	shortCode := generateShortCode()

	// Store in PostgreSQL
	err := storeURL(dbConn, json.LongURL, shortCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store URL"})
		return
	}

	// Respond with shortened URL
	shortURL := fmt.Sprintf("http://localhost:8080/%s", shortCode)
	c.JSON(http.StatusOK, gin.H{"short_url": shortURL})

}

func redirectHandler(c *gin.Context) {
	shortCode := c.Param("shortcode") //Extract Shortcode from URL
	fmt.Printf("Shortcode Extracted: %s", shortCode)

	longURL, err := getLongURL(dbConn, shortCode) //Call getLongURL function to return long url if shortcode matches
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Shortcode not found"}) //If shortcode does not match, return err Not Found
		return
	}

	c.Redirect(http.StatusFound, longURL) // If Shortcode matches, provide http statusfound and redirect to the longURL address

}
