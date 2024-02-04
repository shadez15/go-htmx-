package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/multitemplate"
	"github.com/htmx/htmx-go"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type BlogPost struct {
	gorm.Model
	Title   string
	Content string
}

var db *gorm.DB

func main() {
	// Connect to SQLite database
	setupDatabase()

	// Initialize Gin
	router := gin.Default()

	// Set up HTMX handlers
	setupHTMXHandlers(router)

	// Routes
	router.GET("/", getPosts)
	router.POST("/create", createPost)
	router.GET("/post/:id", getPost)

	// Serve static files
	router.Static("/static", "./static")

	// Run the server
	router.Run(":8080")
}

func setupDatabase() {
	var err error
	db, err = gorm.Open(sqlite.Open("blog.db"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}

	// Auto Migrate the model
	db.AutoMigrate(&BlogPost{})
}

func setupHTMXHandlers(router *gin.Engine) {
	engine := multitemplate.NewEngine()

	// Parse templates
	templatesDir := "templates"
	err := filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".html") {
			_, err = engine.ParseFiles(path)
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		panic(fmt.Sprintf("Error loading templates: %v", err))
	}

	router.HTMLRender = engine

	// Register HTMX middleware
	router.Use(htmx.Middleware(htmx.ConfigureDefaultConfig()))
}

func getPosts(c *gin.Context) {
	var posts []BlogPost
	db.Find(&posts)
	c.HTML(http.StatusOK, "index.html", gin.H{"posts": posts})
}

func createPost(c *gin.Context) {
	var input BlogPost
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create the blog post
	db.Create(&input)

	// Reload the page using HTMX
	c.Header("HX-Trigger", "reload: true")
	getPosts(c)
}

func getPost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var post BlogPost
	result := db.First(&post, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	c.HTML(http.StatusOK, "post.html", gin.H{"post": post})
}