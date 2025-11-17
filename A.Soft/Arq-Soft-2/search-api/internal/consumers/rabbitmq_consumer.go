package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/sporthub/search-api/internal/domain"
	"github.com/sporthub/search-api/internal/repository"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Event struct {
	Op         string `json:"op"` // "create" | "update" | "delete"
	ActivityID string `json:"activityId"`
	SessionID  string `json:"sessionId"`
	Timestamp  string `json:"timestamp"`
}

// Consumer mantiene conexión y dependencias
type Consumer struct {
	repo     *repository.SolrRepo
	cacheL   *repository.LocalCache
	cacheD   *repository.DistCache
	activity string // base URL de activities-api
}

func NewConsumer(solr *repository.SolrRepo, local *repository.LocalCache, dist *repository.DistCache, activitiesAPI string) *Consumer {
	return &Consumer{repo: solr, cacheL: local, cacheD: dist, activity: activitiesAPI}
}

func (c *Consumer) Start(ctx context.Context, conn *amqp.Connection, queue, exchange, routingKey string) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("rabbit channel: %w", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("exchange declare: %w", err)
	}
	q, err := ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("queue declare: %w", err)
	}
	err = ch.QueueBind(q.Name, routingKey, exchange, false, nil)
	if err != nil {
		return fmt.Errorf("queue bind: %w", err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	log.Printf("[consumer] listening for activity events (queue=%s, routingKey=%s)", queue, routingKey)
	go func() {
		for m := range msgs {
			var ev Event
			if err := json.Unmarshal(m.Body, &ev); err != nil {
				log.Printf("[consumer] ERROR: invalid event JSON: %v", err)
				log.Printf("[consumer] Raw message: %s", string(m.Body))
				continue
			}
			log.Printf("[consumer] Received event: op=%s, activityId=%s, sessionId=%s", ev.Op, ev.ActivityID, ev.SessionID)
			c.handle(ctx, ev)
		}
	}()
	<-ctx.Done()
	return nil
}

func (c *Consumer) handle(ctx context.Context, ev Event) {
	log.Printf("[consumer] Processing event: op=%s, activityId=%s, sessionId=%s", ev.Op, ev.ActivityID, ev.SessionID)
	
	// DISEÑO: Solo indexamos actividades en Solr, no sesiones individuales.
	// Las sesiones se indexan como parte de la actividad (usando la primera sesión disponible).
	// Por lo tanto, ignoramos eventos de sesiones individuales.
	if ev.SessionID != "" && ev.SessionID != "0" {
		log.Printf("[consumer] INFO: ignoring session event (sessionId=%s). Only activity events are indexed in Solr.", ev.SessionID)
		return
	}
	
	// Validar que tenemos activityId
	if ev.ActivityID == "" {
		log.Printf("[consumer] WARN: event %s has empty activityId, skipping", ev.Op)
		return
	}
	
	switch ev.Op {
	case "delete":
		log.Printf("[consumer] Deleting activity %s from Solr", ev.ActivityID)
		_ = c.repo.DeleteByID(ctx, ev.ActivityID)
		c.cacheL.Delete(ev.ActivityID)
		c.cacheD.Delete(ev.ActivityID)
		log.Printf("[consumer] SUCCESS: deleted activity %s", ev.ActivityID)
	default: // create/update
		log.Printf("[consumer] Fetching search-doc for activity %s from %s/activities/%s/search-doc", ev.ActivityID, c.activity, ev.ActivityID)
		doc, err := c.fetchActivityDoc(ev.ActivityID)
		if err != nil {
			log.Printf("[consumer] ERROR: fetch error for activity %s: %v", ev.ActivityID, err)
			return
		}
		log.Printf("[consumer] Fetched doc: id=%s, activityId=%s, name=%s", doc.ID, doc.ActivityID, doc.Name)
		log.Printf("[consumer] Indexing activity %s in Solr", ev.ActivityID)
		log.Printf("[consumer] Document to index: id=%s, activityId=%s, name=%s, sport=%s, site=%s, start_dt=%s", 
			doc.ID, doc.ActivityID, doc.Name, doc.Sport, doc.Site, doc.StartAt)
		if err := c.repo.Upsert(ctx, *doc); err != nil {
			log.Printf("[consumer] ERROR: solr upsert error for activity %s: %v", ev.ActivityID, err)
			return
		}
		log.Printf("[consumer] SUCCESS: indexed activity %s (name=%s)", ev.ActivityID, doc.Name)
	}
}

// fetchActivityDoc obtiene el search-doc de una actividad
func (c *Consumer) fetchActivityDoc(activityID string) (*domain.SearchDoc, error) {
	url := fmt.Sprintf("%s/activities/%s/search-doc", c.activity, activityID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("activities-api: %s", string(b))
	}
	var doc domain.SearchDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}
