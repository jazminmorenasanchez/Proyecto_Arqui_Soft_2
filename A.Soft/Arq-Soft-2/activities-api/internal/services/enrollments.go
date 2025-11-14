package services

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/sporthub/activities-api/internal/clients"
	"github.com/sporthub/activities-api/internal/config"
	"github.com/sporthub/activities-api/internal/domain"
	"github.com/sporthub/activities-api/internal/repository"
)

var ErrNoCupo = errors.New("no hay cupo disponible")
var ErrAlreadyEnrolled = errors.New("ya inscripto en esta sesión")
var ErrForbidden = errors.New("forbidden")

type EnrollmentsService struct {
	erepo repository.EnrollmentsRepository
	srepo repository.SessionsRepository
	arepo repository.ActivitiesRepository
	bus   clients.Publisher
	cfg   *config.Config
}

func NewEnrollmentsService(e repository.EnrollmentsRepository, s repository.SessionsRepository, a repository.ActivitiesRepository, bus clients.Publisher, cfg *config.Config) *EnrollmentsService {
	return &EnrollmentsService{erepo: e, srepo: s, arepo: a, bus: bus, cfg: cfg}
}

func (svc *EnrollmentsService) Enroll(ctx context.Context, sessionId uint64, userId string) (uint64, error) {
	// Obtener sesión y actividad
	sess, err := svc.srepo.GetByID(ctx, sessionId)
	if err != nil {
		return 0, err
	}
	act, err := svc.arepo.GetByID(ctx, sess.ActivityID)
	if err != nil {
		return 0, err
	}

	// Check duplicate enrollment
	exists, err := svc.erepo.Exists(ctx, userId, sessionId)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, ErrAlreadyEnrolled
	}

	// Concurrencia: calcular precio final y verificar cupo en paralelo
	type res struct {
		precio float64
		err    error
	}
	priceCh := make(chan res, 1)
	cupoCh := make(chan error, 1)

	wg := sync.WaitGroup{}
	wg.Add(2)

	// Goroutine 1: cálculo de precio (descuentos/promos/horario pico)
	go func(base float64) {
		defer wg.Done()
		precio := base
		// Ejemplos de reglas simuladas:
		// descuento membresía (5%)
		precio = precio * 0.95
		// horario pico: +10% si inicio entre 18:00-22:00
		// (para simplificar, comparamos string "HH:mm")
		if sess.Inicio >= "18:00" && sess.Inicio <= "22:00" {
			precio = precio * 1.10
		}
		priceCh <- res{precio: precio, err: nil}
	}(act.PrecioBase)

	// Goroutine 2: verificación de cupo
	go func() {
		defer wg.Done()
		ocupadas, err := svc.srepo.CountEnrollments(ctx, sessionId)
		if err != nil {
			cupoCh <- err
			return
		}
		if ocupadas >= sess.Capacidad {
			cupoCh <- ErrNoCupo
			return
		}
		cupoCh <- nil
	}()

	wg.Wait()
	close(priceCh)
	close(cupoCh)

	// Recolectar resultados
	var precio float64
	select {
	case r := <-priceCh:
		if r.err != nil {
			return 0, r.err
		}
		precio = r.precio
	case <-time.After(5 * time.Second):
		return 0, errors.New("timeout calculating price")
	}

	select {
	case err := <-cupoCh:
		if err != nil {
			return 0, err
		}
	case <-time.After(5 * time.Second):
		return 0, errors.New("timeout checking capacity")
	}

	// Crear inscripción
	enr := &domain.Enrollment{
		ActivityID:  act.ID,
		SessionID:   sessionId,
		UserID:      userId,
		PrecioFinal: precio,
		Estado:      "confirmada",
		CreatedAt:   time.Now(),
	}
	id, err := svc.erepo.Create(ctx, enr)
	if err != nil {
		return 0, err
	}

	// Publicar evento
	_ = svc.bus.Publish("enrollment.created", map[string]any{
		"op": "enroll", "id": id, "sessionId": sessionId, "activityId": act.ID, "userId": userId, "total": precio, "ts": time.Now(),
	})
	return id, nil
}

func (svc *EnrollmentsService) ListByUser(ctx context.Context, userId string) ([]domain.Enrollment, error) {
	return svc.erepo.ListByUser(ctx, userId)
}

func (svc *EnrollmentsService) CancelEnrollment(ctx context.Context, enrollmentId uint64, requesterUserId string, requesterRole string) error {
	enr, err := svc.erepo.GetByID(ctx, enrollmentId)
	if err != nil {
		return err
	}
	// only owner or admin can cancel
	if requesterRole != "admin" && enr.UserID != requesterUserId {
		return ErrForbidden
	}
	if err := svc.erepo.UpdateStatus(ctx, enrollmentId, "cancelada"); err != nil {
		return err
	}
	_ = svc.bus.Publish("enrollment.cancelled", map[string]any{"op": "cancel", "id": enrollmentId, "sessionId": enr.SessionID, "activityId": enr.ActivityID, "userId": enr.UserID, "ts": time.Now()})
	return nil
}
