package domain

import "time"

type Activity struct {
	ID          uint64    `bson:"_id,omitempty" json:"id"`
	OwnerUserID string    `bson:"ownerUserId"    json:"ownerUserId"`
	Categoria   string    `bson:"categoria"      json:"categoria"`
	Nombre      string    `bson:"nombre"         json:"nombre"`
	Ubicacion   string    `bson:"ubicacion"      json:"ubicacion"`
	Instructor  string    `bson:"instructor"     json:"instructor"`
	PrecioBase  float64   `bson:"precioBase"     json:"precioBase"`
	Rating      float64   `bson:"rating"         json:"rating"`
	UpdatedAt   time.Time `bson:"updatedAt"      json:"updatedAt"`
}
