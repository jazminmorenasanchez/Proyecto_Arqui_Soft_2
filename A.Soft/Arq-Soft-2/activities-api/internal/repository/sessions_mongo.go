package repository

import (
	"context"
	"time"

	"github.com/sporthub/activities-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type SessionsRepository interface {
	Create(ctx context.Context, s *domain.Session) (uint64, error)
	ListByActivity(ctx context.Context, activityId uint64) ([]domain.Session, error)
	GetByID(ctx context.Context, id uint64) (*domain.Session, error)
	Update(ctx context.Context, id uint64, update bson.M) error
	Delete(ctx context.Context, id uint64) error
	CountEnrollments(ctx context.Context, sessionId uint64) (int, error) // helper (via enrollments col)
}

type sessionsMongo struct {
	scol *mongo.Collection
	ecol *mongo.Collection
	db   *mongo.Database
}

func NewSessionsMongo(db *mongo.Database) SessionsRepository {
	return &sessionsMongo{
		scol: db.Collection("sessions"),
		ecol: db.Collection("enrollments"),
		db:   db,
	}
}

func (r *sessionsMongo) Create(ctx context.Context, s *domain.Session) (uint64, error) {
	// Generar ID secuencial
	id, err := getNextSequence(ctx, r.db, "sessions")
	if err != nil {
		return 0, err
	}
	
	s.ID = id
	now := time.Now()
	s.CreatedAt, s.UpdatedAt = now, now
	
	_, err = r.scol.InsertOne(ctx, s)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *sessionsMongo) ListByActivity(ctx context.Context, activityId uint64) ([]domain.Session, error) {
	cur, err := r.scol.Find(ctx, bson.M{"activityId": activityId})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []domain.Session
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *sessionsMongo) GetByID(ctx context.Context, id uint64) (*domain.Session, error) {
	var s domain.Session
	if err := r.scol.FindOne(ctx, bson.M{"_id": id}).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *sessionsMongo) Update(ctx context.Context, id uint64, update bson.M) error {
	update["updatedAt"] = time.Now()
	_, err := r.scol.UpdateByID(ctx, id, bson.M{"$set": update})
	return err
}

func (r *sessionsMongo) Delete(ctx context.Context, id uint64) error {
	_, err := r.scol.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *sessionsMongo) CountEnrollments(ctx context.Context, sessionId uint64) (int, error) {
	n, err := r.ecol.CountDocuments(ctx, bson.M{"sessionId": sessionId, "estado": "confirmada"})
	return int(n), err
}

