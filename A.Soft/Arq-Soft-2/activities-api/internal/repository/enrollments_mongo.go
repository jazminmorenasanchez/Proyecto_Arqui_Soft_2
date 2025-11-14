package repository

import (
	"context"
	"time"

	"github.com/sporthub/activities-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type EnrollmentsRepository interface {
	Create(ctx context.Context, e *domain.Enrollment) (uint64, error)
	ListByUser(ctx context.Context, userId string) ([]domain.Enrollment, error)
	Exists(ctx context.Context, userId string, sessionId uint64) (bool, error)
	GetByID(ctx context.Context, id uint64) (*domain.Enrollment, error)
	UpdateStatus(ctx context.Context, id uint64, status string) error
}

type enrollmentsMongo struct {
	col *mongo.Collection
	db  *mongo.Database
}

func NewEnrollmentsMongo(db *mongo.Database) EnrollmentsRepository {
	return &enrollmentsMongo{
		col: db.Collection("enrollments"),
		db:  db,
	}
}

func (r *enrollmentsMongo) Create(ctx context.Context, e *domain.Enrollment) (uint64, error) {
	// Generar ID secuencial
	id, err := getNextSequence(ctx, r.db, "enrollments")
	if err != nil {
		return 0, err
	}
	
	e.ID = id
	e.CreatedAt = time.Now()
	
	_, err = r.col.InsertOne(ctx, e)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *enrollmentsMongo) ListByUser(ctx context.Context, userId string) ([]domain.Enrollment, error) {
	// Solo devolver inscripciones confirmadas (excluir canceladas)
	cur, err := r.col.Find(ctx, bson.M{"userId": userId, "estado": "confirmada"})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.Enrollment
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *enrollmentsMongo) Exists(ctx context.Context, userId string, sessionId uint64) (bool, error) {
	// Solo contar inscripciones confirmadas, excluir canceladas
	cnt, err := r.col.CountDocuments(ctx, bson.M{
		"userId":    userId,
		"sessionId": sessionId,
		"estado":    "confirmada",
	})
	if err != nil {
		return false, err
	}
	return cnt > 0, nil
}

func (r *enrollmentsMongo) GetByID(ctx context.Context, id uint64) (*domain.Enrollment, error) {
	var out domain.Enrollment
	if err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *enrollmentsMongo) UpdateStatus(ctx context.Context, id uint64, status string) error {
	_, err := r.col.UpdateByID(ctx, id, bson.M{"$set": bson.M{"estado": status}})
	return err
}
