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

type ActivitiesService struct {
	repo  repository.ActivitiesRepository
	users *clients.UsersClient
	bus   *clients.Rabbit
	cfg   *config.Config
}

func NewActivitiesService(r repository.ActivitiesRepository, u *clients.UsersClient, bus *clients.Rabbit, cfg *config.Config) *ActivitiesService {
	return &ActivitiesService{repo: r, users: u, bus: bus, cfg: cfg}
}

func (s *ActivitiesService) Create(ctx context.Context, a *domain.Activity) (uint64, error) {
	if _, err := s.users.GetUser(a.OwnerUserID); err != nil {
		return 0, err
	}
	id, err := s.repo.Create(ctx, a)
	if err != nil {
		return 0, err
	}
	// Publicar evento para search-api con routing key que coincida con los bindings de RabbitMQ
	_ = s.bus.Publish("activity.created", map[string]any{
		"op":         "create",
		"activityId": fmt.Sprintf("%d", id),
		"sessionId":  "", // Vacío porque es un evento de actividad
		"timestamp":  time.Now().Format(time.RFC3339),
	})
	return id, nil
}

func (s *ActivitiesService) GetByID(ctx context.Context, id uint64) (*domain.Activity, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ActivitiesService) Update(ctx context.Context, id uint64, update bson.M) error {
	if err := s.repo.Update(ctx, id, update); err != nil {
		return err
	}
	// Publicar evento para search-api con routing key que coincida con los bindings de RabbitMQ
	_ = s.bus.Publish("activity.updated", map[string]any{
		"op":         "update",
		"activityId": fmt.Sprintf("%d", id),
		"sessionId":  "", // Vacío porque es un evento de actividad
		"timestamp":  time.Now().Format(time.RFC3339),
	})
	return nil
}

func (s *ActivitiesService) Delete(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	// Publicar evento para search-api con routing key que coincida con los bindings de RabbitMQ
	_ = s.bus.Publish("activity.deleted", map[string]any{
		"op":         "delete",
		"activityId": fmt.Sprintf("%d", id),
		"sessionId":  "", // Vacío porque es un evento de actividad
		"timestamp":  time.Now().Format(time.RFC3339),
	})
	return nil
}

func (s *ActivitiesService) List(ctx context.Context, skip int, limit int) ([]*domain.Activity, int64, error) {
	return s.repo.List(ctx, skip, limit)
}

// ReindexAll publica eventos de actualización para todas las actividades existentes
// Esto hace que el consumer de search-api las indexe en Solr
func (s *ActivitiesService) ReindexAll(ctx context.Context) (int, error) {
	// Obtener todas las actividades (sin límite)
	activities, _, err := s.repo.List(ctx, 0, 10000) // Usar un límite alto para obtener todas
	if err != nil {
		return 0, err
	}
	
	count := 0
	for _, activity := range activities {
		// Publicar evento de actualización para cada actividad
		_ = s.bus.Publish("activity.updated", map[string]any{
			"op":         "update",
			"activityId": fmt.Sprintf("%d", activity.ID),
			"sessionId":  "", // Vacío porque es un evento de actividad
			"timestamp":  time.Now().Format(time.RFC3339),
		})
		count++
	}
	return count, nil
}
