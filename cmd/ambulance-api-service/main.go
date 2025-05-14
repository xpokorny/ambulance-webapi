package main

import (
    "log"
    "os"
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/xpokorny/ambulance-webapi/api"
	"github.com/xpokorny/ambulance-webapi/internal/ambulance_wl"
	"github.com/xpokorny/ambulance-webapi/internal/db_service"
    "context"
    "time"
    "github.com/gin-contrib/cors"
)

func initializeTestData(dbService db_service.DbService[ambulance_wl.User], locationService db_service.DbService[ambulance_wl.Location]) {
    // Check if data already exists by trying to find a known user
    _, err := dbService.FindDocument(context.Background(), "user1")
    if err == nil {
        log.Printf("Database already initialized, skipping test data creation")
        return
    }

    log.Printf("Initializing database with test data...")
    
    // Create test users
    users := []ambulance_wl.User{
        {
            Id:   "user1",
            Name: "John Doe",
            Role: "patient",
        },
        {
            Id:   "user2",
            Name: "Jane Smith",
            Role: "patient",
        },
        {
            Id:   "user3",
            Name: "Robert Johnson",
            Role: "patient",
        },
        {
            Id:   "user4",
            Name: "Emily Davis",
            Role: "patient",
        },
        {
            Id:   "user5",
            Name: "Dr. Michael Brown",
            Role: "doctor",
        },
        {
            Id:   "user6",
            Name: "Dr. Sarah Wilson",
            Role: "doctor",
        },
        {
            Id:   "user7",
            Name: "Dr. James Miller",
            Role: "doctor",
        },
        {
            Id:   "user8",
            Name: "Dr. Lisa Taylor",
            Role: "doctor",
        },
    }

    for _, user := range users {
        err := dbService.CreateDocument(context.Background(), user.Id, &user)
        if err != nil && err != db_service.ErrConflict {
            log.Printf("Failed to create user %s: %v", user.Id, err)
        }
    }

    // Create test locations
    locations := []ambulance_wl.Location{
        {
            Id:      "loc1",
            Name:    "Room 101",
            Address: "Main Building, Floor 1",
        },
        {
            Id:      "loc2",
            Name:    "Room 202",
            Address: "Main Building, Floor 2",
        },
        {
            Id:      "loc3",
            Name:    "Room 303",
            Address: "Main Building, Floor 3",
        },
    }

    for _, location := range locations {
        err := locationService.CreateDocument(context.Background(), location.Id, &location)
        if err != nil && err != db_service.ErrConflict {
            log.Printf("Failed to create location %s: %v", location.Id, err)
        }
    }

    log.Printf("Database initialization completed")
}

func main() {
    log.Printf("Server started")
    port := os.Getenv("AMBULANCE_API_PORT")
    if port == "" {
        port = "8080"
    }
    environment := os.Getenv("AMBULANCE_API_ENVIRONMENT")
    if !strings.EqualFold(environment, "production") { // case insensitive comparison
        gin.SetMode(gin.DebugMode)
    }
    engine := gin.New()
    engine.Use(gin.Recovery())
	corsMiddleware := cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH"},
        AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
        ExposeHeaders:    []string{""},
        AllowCredentials: false,
        MaxAge: 12 * time.Hour,
    })
    engine.Use(corsMiddleware)

    // setup context update middleware
    appointmentService := db_service.NewMongoService[ambulance_wl.Appointment](db_service.MongoServiceConfig{
        Collection: "appointments",
    })
    userService := db_service.NewMongoService[ambulance_wl.User](db_service.MongoServiceConfig{
        Collection: "users",
    })
    locationService := db_service.NewMongoService[ambulance_wl.Location](db_service.MongoServiceConfig{
        Collection: "locations",
    })
    defer appointmentService.Disconnect(context.Background())
    defer userService.Disconnect(context.Background())
    defer locationService.Disconnect(context.Background())

    // Initialize test data
    initializeTestData(userService, locationService)

    engine.Use(func(ctx *gin.Context) {
        ctx.Set("appointment_service", appointmentService)
        ctx.Set("user_service", userService)
        ctx.Set("location_service", locationService)
        ctx.Next()
    })

    // request routings
	handleFunctions := &ambulance_wl.ApiHandleFunctions{
		AppointmentsAPI: ambulance_wl.NewAppointmentsAPI(),
		UsersAPI:        ambulance_wl.NewUsersAPI(),
		LocationsAPI:    ambulance_wl.NewLocationsAPI(),
	}
	ambulance_wl.NewRouterWithGinEngine(engine, *handleFunctions)
    engine.GET("/openapi", api.HandleOpenApi)
    engine.Run(":" + port)
}