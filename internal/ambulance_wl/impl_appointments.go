package ambulance_wl

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/xpokorny/ambulance-webapi/internal/db_service"
)

type implAppointmentsAPI struct {
}

func NewAppointmentsAPI() AppointmentsAPI {
    return &implAppointmentsAPI{}
}

func (o implAppointmentsAPI) CreateAppointment(c *gin.Context) {
    // get db service from context
    value, exists := c.Get("appointment_service")
    if !exists {
        c.JSON(
            http.StatusInternalServerError,
            gin.H{
                "status":  "Internal Server Error",
                "message": "appointment_service not found",
                "error":   "appointment_service not found",
            })
        return
    }

    db, ok := value.(db_service.DbService[Appointment])
    if !ok {
        c.JSON(
            http.StatusInternalServerError,
            gin.H{
                "status":  "Internal Server Error",
                "message": "appointment_service context is not of required type",
                "error":   "cannot cast appointment_service context to db_service.DbService",
            })
        return
    }

    appointment := Appointment{}
    err := c.BindJSON(&appointment)
    if err != nil {
        c.JSON(
            http.StatusBadRequest,
            gin.H{
                "status":  "Bad Request",
                "message": "Invalid request body",
                "error":   err.Error(),
            })
        return
    }

    if appointment.Id == "" {
        appointment.Id = uuid.New().String()
    }

    err = db.CreateDocument(c, appointment.Id, &appointment)

    switch err {
    case nil:
        c.JSON(
            http.StatusCreated,
            appointment,
        )
    case db_service.ErrConflict:
        c.JSON(
            http.StatusConflict,
            gin.H{
                "status":  "Conflict",
                "message": "Appointment already exists",
                "error":   err.Error(),
            },
        )
    default:
        c.JSON(
            http.StatusBadGateway,
            gin.H{
                "status":  "Bad Gateway",
                "message": "Failed to create appointment in database",
                "error":   err.Error(),
            },
        )
    }
}

func (o implAppointmentsAPI) DeleteAppointment(c *gin.Context) {
    // get db service from context
    value, exists := c.Get("appointment_service")
    if !exists {
        c.JSON(
            http.StatusInternalServerError,
            gin.H{
                "status":  "Internal Server Error",
                "message": "appointment_service not found",
                "error":   "appointment_service not found",
            })
        return
    }

    db, ok := value.(db_service.DbService[Appointment])
    if !ok {
        c.JSON(
            http.StatusInternalServerError,
            gin.H{
                "status":  "Internal Server Error",
                "message": "appointment_service context is not of type db_service.DbService",
                "error":   "cannot cast appointment_service context to db_service.DbService",
            })
        return
    }

    appointmentId := c.Param("appointmentId")
    err := db.DeleteDocument(c, appointmentId)

    switch err {
    case nil:
        c.AbortWithStatus(http.StatusNoContent)
    case db_service.ErrNotFound:
        c.JSON(
            http.StatusNotFound,
            gin.H{
                "status":  "Not Found",
                "message": "Appointment not found",
                "error":   err.Error(),
            },
        )
    default:
        c.JSON(
            http.StatusBadGateway,
            gin.H{
                "status":  "Bad Gateway",
                "message": "Failed to delete appointment from database",
                "error":   err.Error(),
            })
    }
}

func (o implAppointmentsAPI) GetAppointment(c *gin.Context) {
    // get db service from context
    value, exists := c.Get("appointment_service")
    if !exists {
        c.JSON(
            http.StatusInternalServerError,
            gin.H{
                "status":  "Internal Server Error",
                "message": "appointment_service not found",
                "error":   "appointment_service not found",
            })
        return
    }

    db, ok := value.(db_service.DbService[Appointment])
    if !ok {
        c.JSON(
            http.StatusInternalServerError,
            gin.H{
                "status":  "Internal Server Error",
                "message": "appointment_service context is not of type db_service.DbService",
                "error":   "cannot cast appointment_service context to db_service.DbService",
            })
        return
    }

    appointmentId := c.Param("appointmentId")
    if appointmentId == "" {
        c.JSON(
            http.StatusBadRequest,
            gin.H{
                "status":  "Bad Request",
                "message": "Appointment ID is required",
            })
        return
    }

    appointment, err := db.FindDocument(c, appointmentId)
    if err != nil {
        switch err {
        case db_service.ErrNotFound:
            c.JSON(
                http.StatusNotFound,
                gin.H{
                    "status":  "Not Found",
                    "message": "Appointment not found",
                    "error":   err.Error(),
                })
        default:
            c.JSON(
                http.StatusBadGateway,
                gin.H{
                    "status":  "Bad Gateway",
                    "message": "Failed to get appointment from database",
                    "error":   err.Error(),
                })
        }
        return
    }

    c.JSON(http.StatusOK, appointment)
}

func (o implAppointmentsAPI) GetAppointments(c *gin.Context) {
    // get db service from context
    value, exists := c.Get("appointment_service")
    if !exists {
        c.JSON(
            http.StatusInternalServerError,
            gin.H{
                "status":  "Internal Server Error",
                "message": "appointment_service not found",
                "error":   "appointment_service not found",
            })
        return
    }

    db, ok := value.(db_service.DbService[Appointment])
    if !ok {
        c.JSON(
            http.StatusInternalServerError,
            gin.H{
                "status":  "Internal Server Error",
                "message": "appointment_service context is not of type db_service.DbService",
                "error":   "cannot cast appointment_service context to db_service.DbService",
            })
        return
    }

    appointments := []Appointment{}
    cursor, err := db.FindAllDocuments(c.Request.Context())
    if err != nil {
        c.JSON(
            http.StatusInternalServerError,
            gin.H{
                "error":   "failed to get appointments",
                "message": err.Error(),
                "status":  "Internal Server Error",
            })
        return
    }
    defer cursor.Close(c.Request.Context())

    for cursor.Next(c.Request.Context()) {
        var appointment Appointment
        if err := cursor.Decode(&appointment); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "failed to decode appointment",
                "message": err.Error(),
                "status":  "Internal Server Error",
            })
            return
        }
        appointments = append(appointments, appointment)
    }

    userId := c.Query("userId")
    role := c.Query("role")

    if userId != "" || role != "" {
        filteredAppointments := []Appointment{}
        for _, appointment := range appointments {
            matches := false

            if userId != "" {
                if appointment.Patient.Id == userId || 
                   appointment.Doctor.Id == userId || 
                   appointment.CreatedBy.Id == userId {
                    matches = true
                }
            }

            if role != "" {
                switch role {
                case "patient":
                    if appointment.Patient.Id == userId {
                        matches = true
                    }
                case "creator":
                    if appointment.CreatedBy.Id == userId {
                        matches = true
                    }
                }
            }

            if (userId == "" && role == "") || matches {
                filteredAppointments = append(filteredAppointments, appointment)
            }
        }
        appointments = filteredAppointments
    }

    c.JSON(http.StatusOK, appointments)
}

func (o implAppointmentsAPI) UpdateAppointment(c *gin.Context) {
    // get db service from context
    value, exists := c.Get("appointment_service")
    if !exists {
        c.JSON(
            http.StatusInternalServerError,
            gin.H{
                "status":  "Internal Server Error",
                "message": "appointment_service not found",
                "error":   "appointment_service not found",
            })
        return
    }

    db, ok := value.(db_service.DbService[Appointment])
    if !ok {
        c.JSON(
            http.StatusInternalServerError,
            gin.H{
                "status":  "Internal Server Error",
                "message": "appointment_service context is not of type db_service.DbService",
                "error":   "cannot cast appointment_service context to db_service.DbService",
            })
        return
    }

    appointmentId := c.Param("appointmentId")
    if appointmentId == "" {
        c.JSON(
            http.StatusBadRequest,
            gin.H{
                "status":  "Bad Request",
                "message": "Appointment ID is required",
            })
        return
    }

    existingAppointment, err := db.FindDocument(c, appointmentId)
    if err != nil {
        switch err {
        case db_service.ErrNotFound:
            c.JSON(
                http.StatusNotFound,
                gin.H{
                    "status":  "Not Found",
                    "message": "Appointment not found",
                    "error":   err.Error(),
                })
        default:
            c.JSON(
                http.StatusBadGateway,
                gin.H{
                    "status":  "Bad Gateway",
                    "message": "Failed to get appointment from database",
                    "error":   err.Error(),
                })
        }
        return
    }

    var updateData Appointment
    if err := c.ShouldBindJSON(&updateData); err != nil {
        c.JSON(
            http.StatusBadRequest,
            gin.H{
                "status":  "Bad Request",
                "message": "Invalid request body",
                "error":   err.Error(),
            })
        return
    }

    if updateData.Id != "" {
        existingAppointment.Id = updateData.Id
    }
    if !updateData.DateTime.IsZero() {
        existingAppointment.DateTime = updateData.DateTime
    }

    if updateData.Patient.Id != "" {
        existingAppointment.Patient.Id = updateData.Patient.Id
    }
    if updateData.Patient.Name != "" {
        existingAppointment.Patient.Name = updateData.Patient.Name
    }
    if updateData.Patient.Role != "" {
        existingAppointment.Patient.Role = updateData.Patient.Role
    }

    if updateData.Doctor.Id != "" {
        existingAppointment.Doctor.Id = updateData.Doctor.Id
    }
    if updateData.Doctor.Name != "" {
        existingAppointment.Doctor.Name = updateData.Doctor.Name
    }
    if updateData.Doctor.Role != "" {
        existingAppointment.Doctor.Role = updateData.Doctor.Role
    }

    if updateData.Location.Id != "" {
        existingAppointment.Location.Id = updateData.Location.Id
    }
    if updateData.Location.Name != "" {
        existingAppointment.Location.Name = updateData.Location.Name
    }
    if updateData.Location.Address != "" {
        existingAppointment.Location.Address = updateData.Location.Address
    }

    if updateData.CreatedBy.Id != "" {
        existingAppointment.CreatedBy.Id = updateData.CreatedBy.Id
    }
    if updateData.CreatedBy.Name != "" {
        existingAppointment.CreatedBy.Name = updateData.CreatedBy.Name
    }
    if updateData.CreatedBy.Role != "" {
        existingAppointment.CreatedBy.Role = updateData.CreatedBy.Role
    }

    err = db.UpdateDocument(c, appointmentId, existingAppointment)
    if err != nil {
        c.JSON(
            http.StatusBadGateway,
            gin.H{
                "status":  "Bad Gateway",
                "message": "Failed to update appointment in database",
                "error":   err.Error(),
            })
        return
    }

    c.JSON(http.StatusOK, existingAppointment)
}