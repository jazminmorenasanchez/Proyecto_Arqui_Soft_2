package controllers

import "time"

type CreateActivityRequest struct {
	Categoria  string  `json:"categoria" binding:"required"`
	Nombre     string  `json:"nombre" binding:"required"`
	Ubicacion  string  `json:"ubicacion" binding:"required"`
	Instructor string  `json:"instructor"`
	PrecioBase float64 `json:"precioBase" binding:"required,gt=0"`
}

type CreateSessionRequest struct {
	ActivityID string    `json:"activityId" binding:"required"`
	StartTime  time.Time `json:"startTime" binding:"required"`
	EndTime    time.Time `json:"endTime" binding:"required,gtfield=StartTime"`
	Capacity   int       `json:"capacity" binding:"required,gt=0"`
}

type CreateEnrollmentRequest struct {
	SessionID string `json:"sessionId" binding:"required"`
}

type PaginationQuery struct {
	Limit int `form:"limit,default=10" binding:"min=1,max=100"`
	Skip  int `form:"skip,default=0" binding:"min=0"`
}
