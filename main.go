package main

import (
	"embed"
	"fmt"
	"log"
	"net/http" // Add this import
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	//"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

//go:embed all:frontend/dist
var assets embed.FS
// Declare fiberApp globally so it can be used across functions
var fiberApp *fiber.App

func main() {
    // Load .env file if it exists
    err := godotenv.Load()
    if err != nil {
        fmt.Println("No .env file found, using default settings")
    }

    app := NewApp()

    // Check if we should run in server mode
    serverMode := os.Getenv("SERVER_MODE")
    //devMode := os.Getenv("DEV_MODE")
    if serverMode == "true" {
        fiberApp = fiber.New()


        // Enable CORS for all routes and origins
        fiberApp.Use(cors.New(cors.Config{
            AllowOrigins: "*", // Allow all origins
            AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
        }))

        fmt.Println("Running server mode")
        runServerMode(app)
    }

    runDesktopMode(app)
}



func runServerMode(app *App) {
    
    // Serve the React frontend from the embedded assets
    fiberApp.Use("/", filesystem.New(filesystem.Config{
        Root:       http.FS(assets), // Serve embedded assets
        PathPrefix: "frontend/dist", // Path within the embedded assets
        Index:      "index.html",    // Serve index.html for the root path
    }))

    // Set up API routes
    setupRoutes(fiberApp, app)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    host := os.Getenv("SERVER_HOST")
    if host == "" {
        host = "localhost"
    }

    fmt.Printf("Server running on http://%s:%s\n", host, port)
    fiberApp.Listen(fmt.Sprintf("%s:%s", host, port))
}

func runDesktopMode(app *App) {
    defer func() {
        if r := recover(); r != nil {
            log.Println("Desktop GUI not loaded. Error:", r)
            runtime.LogErrorf(app.ctx, "Desktop GUI not loaded: %v", r)
        }
    }()

    // Create application with options
    err := wails.Run(&options.App{
        Title:  "NeuralForge",
        Width:  1024,
        Height: 768,
        AssetServer: &assetserver.Options{
            Assets: assets,
        },
        BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
        OnStartup:        app.startup,
        Bind: []interface{}{
            app,
        },
        //StartHidden:       true, // Keep the window hidden on startup
        //HideWindowOnClose: true, // Hide the window instead of closing
        /*OnDomReady: func(ctx context.Context) {
            // Logic to execute when the DOM is ready
        },
        OnBeforeClose: func(ctx context.Context) (prevent bool) {
            runtime.WindowHide(ctx) // Use Wails' runtime to hide the window
            return true // Prevent the window from closing
        },*/
    })

    if err != nil {
        log.Println("Desktop GUI not loaded. Error:", err.Error())
        runtime.LogErrorf(app.ctx, "Desktop GUI not loaded: %v", err.Error())
    }
}

func setupRoutes(fiberApp *fiber.App, appLogic *App) {
    // Greet route
    fiberApp.Get("/api/greet/:name", func(c *fiber.Ctx) error {
        name := c.Params("name")
        if name == "" {
            return c.Status(400).SendString("Name parameter is required")
        }
        greeting := appLogic.Greet(name)
        return c.Status(200).SendString(greeting) // Send 200 status for success
    })

    // List projects route
    fiberApp.Get("/api/list-projects", func(c *fiber.Ctx) error {
        projects, err := appLogic.ListProjects()
        if err != nil {
            return c.Status(500).SendString("Failed to list projects: " + err.Error())
        }
        if len(projects) == 0 {
            return c.Status(204).JSON([]string{}) // Send 204 No Content if no projects found
        }
        return c.Status(200).JSON(projects) // Send 200 status for success with project list
    })

    // Create project route
    fiberApp.Post("/api/create-project", func(c *fiber.Ctx) error {
        var body struct {
            ProjectName string `json:"projectName"`
        }
        if err := c.BodyParser(&body); err != nil {
            return c.Status(400).SendString("Invalid request body: " + err.Error())
        }
        if body.ProjectName == "" {
            return c.Status(400).SendString("ProjectName is required")
        }
        projectDir, err := appLogic.CreateProject(body.ProjectName)
        if err != nil {
            return c.Status(500).SendString("Failed to create project: " + err.Error())
        }
        return c.Status(201).SendString(projectDir) // Send 201 status for successful creation
    })

    // Add more routes as needed
}
