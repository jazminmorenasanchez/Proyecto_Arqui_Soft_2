package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/sporthub/activities-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ActivitiesRepository interface {
	Create(ctx context.Context, a *domain.Activity) (uint64, error)
	GetByID(ctx context.Context, id uint64) (*domain.Activity, error)
	Update(ctx context.Context, id uint64, update bson.M) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context, skip int, limit int) ([]*domain.Activity, int64, error)
}

type activitiesMongo struct {
	col *mongo.Collection
	db  *mongo.Database
}

func NewActivitiesMongo(db *mongo.Database) ActivitiesRepository {
	return &activitiesMongo{
		col: db.Collection("activities"),
		db:  db,
	}
}

func (r *activitiesMongo) Create(ctx context.Context, a *domain.Activity) (uint64, error) {
	// Generar ID secuencial
	id, err := getNextSequence(ctx, r.db, "activities")
	if err != nil {
		return 0, err
	}
	
	a.ID = id
	a.UpdatedAt = time.Now()
	
	_, err = r.col.InsertOne(ctx, a)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *activitiesMongo) GetByID(ctx context.Context, id uint64) (*domain.Activity, error) {
	var out domain.Activity
	if err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *activitiesMongo) Update(ctx context.Context, id uint64, update bson.M) error {
	update["updatedAt"] = time.Now()
	_, err := r.col.UpdateByID(ctx, id, bson.M{"$set": update})
	return err
}

func (r *activitiesMongo) Delete(ctx context.Context, id uint64) error {
	_, err := r.col.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *activitiesMongo) List(ctx context.Context, skip int, limit int) ([]*domain.Activity, int64, error) {
	// Si no hay límite especificado o es 0, usar default
	if limit <= 0 {
		limit = 10
	}
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.M{"updatedAt": -1})

	total, err := r.col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	cur, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cur.Close(ctx)

	var out []*domain.Activity
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, fmt.Errorf("failed to decode documents: %w", err)
	}
	return out, total, nil
}
