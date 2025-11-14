package domain

import "time"

type Enrollment struct {
	ID          uint64    `bson:"_id,omitempty" json:"id"`
	ActivityID  uint64    `bson:"activityId"    json:"activityId"`
	SessionID   uint64    `bson:"sessionId"     json:"sessionId"`
	UserID      string    `bson:"userId"        json:"userId"`
	PrecioFinal float64   `bson:"precioFinal"   json:"precioFinal"`
	Estado      string    `bson:"estado"        json:"estado"` // pendiente|confirmada|cancelada
	CreatedAt   time.Time `bson:"createdAt"     json:"createdAt"`
}
