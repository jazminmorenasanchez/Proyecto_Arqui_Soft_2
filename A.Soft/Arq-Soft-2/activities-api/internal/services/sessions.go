package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sporthub/activities-api/internal/clients"
	"github.com/sporthub/activities-api/internal/config"
	"github.com/sporthub/activities-api/internal/domain"
	"github.com/sporthub/activities-api/internal/repository"
	"go.mongodb.org/mongo-driver/bson"
)

type SessionsService struct {
	srepo repository.SessionsRepository
	arepo repository.ActivitiesRepository
	bus   *clients.Rabbit
	cfg   *config.Config
}

func NewSessionsService(s repository.SessionsRepository, a repository.ActivitiesRepository, bus *clients.Rabbit, cfg *config.Config) *SessionsService {
	return &SessionsService{srepo: s, arepo: a, bus: bus, cfg: cfg}
}

func (s *SessionsService) Create(ctx context.Context, sess *domain.Session) (uint64, error) {
	// valida que exista la actividad
	if _, err := s.arepo.GetByID(ctx, sess.ActivityID); err != nil {
		return 0, err
	}
	id, err := s.srepo.Create(ctx, sess)
	if err != nil {
		return 0, err
	}
	_ = s.bus.Publish("activity.session.created", map[string]any{
		"op":         "create",
		"sessionId":  fmt.Sprintf("%d", id),
		"activityId": fmt.Sprintf("%d", sess.ActivityID),
		"timestamp":  time.Now().Format(time.RFC3339),
	})
	return id, nil
}

func (s *SessionsService) ListByActivity(ctx context.Context, activityId uint64) ([]domain.Session, error) {
	return s.srepo.ListByActivity(ctx, activityId)
}

func (s *SessionsService) Update(ctx context.Context, id uint64, update bson.M) error {
	if err := s.srepo.Update(ctx, id, update); err != nil {
		return err
	}
	// Obtener la sesión para tener el activityId
	session, err := s.srepo.GetByID(ctx, id)
	if err == nil {
		_ = s.bus.Publish("activity.session.updated", map[string]any{
			"op":         "update",
			"sessionId":  fmt.Sprintf("%d", id),
			"activityId": fmt.Sprintf("%d", session.ActivityID),
			"timestamp":  time.Now().Format(time.RFC3339),
		})
	}
	return nil
}

func (s *SessionsService) Delete(ctx context.Context, id uint64) error {
	// Obtener la sesión para tener el activityId antes de eliminar
	session, err := s.srepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	if err := s.srepo.Delete(ctx, id); err != nil {
		return err
	}
	
	_ = s.bus.Publish("activity.session.deleted", map[string]any{
		"op":         "delete",
		"sessionId":  fmt.Sprintf("%d", id),
		"activityId": fmt.Sprintf("%d", session.ActivityID),
		"timestamp":  time.Now().Format(time.RFC3339),
	})
	return nil
}

func (s *SessionsService) GetByID(ctx context.Context, id uint64) (*domain.Session, error) {
	return s.srepo.GetByID(ctx, id)
}

func (s *SessionsService) UpdateSession(ctx context.Context, id uint64, update *domain.Session) error {
	updateMap := bson.M{
		"fecha":     update.Fecha,
		"inicio":    update.Inicio,
		"fin":       update.Fin,
		"capacidad": update.Capacidad,
		"updatedAt": time.Now(),
	}
	return s.Update(ctx, id, updateMap)
}

func (s *SessionsService) CreateSession(ctx context.Context, sess *domain.Session) (uint64, error) {
	return s.Create(ctx, sess)
}

func (s *SessionsService) GetSessionsByActivity(ctx context.Context, activityId uint64) ([]domain.Session, error) {
	return s.ListByActivity(ctx, activityId)
}

func (s *SessionsService) GetSessionByID(ctx context.Context, id uint64) (*domain.Session, error) {
	return s.GetByID(ctx, id)
}

func (s *SessionsService) DeleteSession(ctx context.Context, id uint64) error {
	return s.Delete(ctx, id)
}
