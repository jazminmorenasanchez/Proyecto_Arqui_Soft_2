package domain

import "time"

type Session struct {
	ID         uint64    `bson:"_id,omitempty" json:"id"`
	ActivityID uint64    `bson:"activityId"    json:"activityId"`
	Fecha      string    `bson:"fecha"         json:"fecha"`  // YYYY-MM-DD
	Inicio     string    `bson:"inicio"        json:"inicio"` // HH:mm
	Fin        string    `bson:"fin"           json:"fin"`    // HH:mm
	Capacidad  int       `bson:"capacidad"     json:"capacidad"`
	CreatedAt  time.Time `bson:"createdAt"     json:"createdAt"`
	UpdatedAt  time.Time `bson:"updatedAt"     json:"updatedAt"`
}
