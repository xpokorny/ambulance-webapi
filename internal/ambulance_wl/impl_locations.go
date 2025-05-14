package ambulance_wl

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/xpokorny/ambulance-webapi/internal/db_service"
)

type implLocationsAPI struct {
}

func NewLocationsAPI() LocationsAPI {
    return &implLocationsAPI{}
}

func (api *implLocationsAPI) GetLocations(c *gin.Context) {
    value, exists := c.Get("location_service")
    if !exists {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "location_service not found",
            "message": "location_service context is not set",
            "status":  "Internal Server Error",
        })
        return
    }

    locationService, ok := value.(db_service.DbService[Location])
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "cannot cast location_service context",
            "message": "location_service context is not of type db_service.DbService",
            "status":  "Internal Server Error",
        })
        return
    }

    // Get all locations
    locations := []Location{}
    cursor, err := locationService.FindAllDocuments(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "failed to get locations",
            "message": err.Error(),
            "status":  "Internal Server Error",
        })
        return
    }
    defer cursor.Close(c.Request.Context())

    for cursor.Next(c.Request.Context()) {
        var location Location
        if err := cursor.Decode(&location); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "failed to decode location",
                "message": err.Error(),
                "status":  "Internal Server Error",
            })
            return
        }
        locations = append(locations, location)
    }

    c.JSON(http.StatusOK, locations)
}