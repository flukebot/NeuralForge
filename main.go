package main

import (
    "embed"
    "fmt"
    "log"
    "os"

    "github.com/gofiber/fiber/v2"
    "github.com/joho/godotenv"
    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    // Load .env file if it exists
    err := godotenv.Load()
    if err != nil {
        fmt.Println("No .env file found, using default settings")
    }

    app := NewApp()

    // Check if we should run in server mode
    serverMode := os.Getenv("SERVER_MODE")
    if serverMode == "true" {
        runServerMode(app)
    }else{
		runDesktopMode(app)
	}

    
}

func runServerMode(app *App) {
    fiberApp := fiber.New()

    // Set up routes
    setupRoutes(fiberApp, app)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    fmt.Printf("Server running on http://localhost:%s\n", port)
    log.Fatal(fiberApp.Listen(":" + port))
}

func runDesktopMode(app *App) {
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
    })

    if err != nil {
        println("Error:", err.Error())
    }
}

func setupRoutes(fiberApp *fiber.App, appLogic *App) {
    fiberApp.Get("/api/greet/:name", func(c *fiber.Ctx) error {
        name := c.Params("name")
        return c.SendString(appLogic.Greet(name))
    })

    fiberApp.Get("/api/list-projects", func(c *fiber.Ctx) error {
        projects, err := appLogic.ListProjects()
        if err != nil {
            return c.Status(500).SendString(err.Error())
        }
        return c.JSON(projects)
    })

    fiberApp.Post("/api/create-project", func(c *fiber.Ctx) error {
        var body struct {
            ProjectName string `json:"projectName"`
        }
        if err := c.BodyParser(&body); err != nil {
            return c.Status(400).SendString(err.Error())
        }
        projectDir, err := appLogic.CreateProject(body.ProjectName)
        if err != nil {
            return c.Status(500).SendString(err.Error())
        }
        return c.SendString(projectDir)
    })

    // Add more routes as needed
}
