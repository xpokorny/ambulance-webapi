package ambulance_wl

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/xpokorny/ambulance-webapi/internal/db_service"
)

type implUsersAPI struct {
}

func NewUsersAPI() UsersAPI {
    return &implUsersAPI{}
}

func (api *implUsersAPI) GetUsers(c *gin.Context) {
    value, exists := c.Get("user_service")
    if !exists {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "user_service not found",
            "message": "user_service context is not set",
            "status":  "Internal Server Error",
        })
        return
    }

    userService, ok := value.(db_service.DbService[User])
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "cannot cast user_service context",
            "message": "user_service context is not of type db_service.DbService",
            "status":  "Internal Server Error",
        })
        return
    }

    // Get all users
    users := []User{}
    cursor, err := userService.FindAllDocuments(c.Request.Context())
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "failed to get users",
            "message": err.Error(),
            "status":  "Internal Server Error",
        })
        return
    }
    defer cursor.Close(c.Request.Context())

    for cursor.Next(c.Request.Context()) {
        var user User
        if err := cursor.Decode(&user); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "failed to decode user",
                "message": err.Error(),
                "status":  "Internal Server Error",
            })
            return
        }
        users = append(users, user)
    }

    // Filter by role if specified
    role := c.Query("role")
    if role != "" {
        filteredUsers := make([]User, 0)
        for _, user := range users {
            if user.Role == role {
                filteredUsers = append(filteredUsers, user)
            }
        }
        users = filteredUsers
    }

    c.JSON(http.StatusOK, users)
}

func (api *implUsersAPI) GetUser(c *gin.Context) {
    userId := c.Param("userId")
    if userId == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "missing user id",
            "message": "user id is required",
            "status":  "Bad Request",
        })
        return
    }

    value, exists := c.Get("user_service")
    if !exists {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "user_service not found",
            "message": "user_service context is not set",
            "status":  "Internal Server Error",
        })
        return
    }

    userService, ok := value.(db_service.DbService[User])
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "cannot cast user_service context",
            "message": "user_service context is not of type db_service.DbService",
            "status":  "Internal Server Error",
        })
        return
    }

    user, err := userService.FindDocument(c.Request.Context(), userId)
    if err != nil {
        if err == db_service.ErrNotFound {
            c.JSON(http.StatusNotFound, gin.H{
                "error":   "user not found",
                "message": "user with id " + userId + " not found",
                "status":  "Not Found",
            })
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "failed to get user",
            "message": err.Error(),
            "status":  "Internal Server Error",
        })
        return
    }

    c.JSON(http.StatusOK, user)
}