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
	
	// Validar que SessionID no esté vacío
	if ev.SessionID == "" {
		log.Printf("[consumer] WARN: event %s has empty sessionId, skipping", ev.Op)
		return
	}
	
	switch ev.Op {
	case "delete":
		log.Printf("[consumer] Deleting session %s from Solr", ev.SessionID)
		_ = c.repo.DeleteByID(ctx, ev.SessionID)
		c.cacheL.Delete(ev.SessionID)
		c.cacheD.Delete(ev.SessionID)
		log.Printf("[consumer] SUCCESS: deleted session %s", ev.SessionID)
	default: // create/update
		log.Printf("[consumer] Fetching search-doc for session %s from %s/sessions/%s/search-doc", ev.SessionID, c.activity, ev.SessionID)
		doc, err := c.fetchDoc(ev.SessionID)
		if err != nil {
			log.Printf("[consumer] ERROR: fetch error for session %s: %v", ev.SessionID, err)
			return
		}
		log.Printf("[consumer] Fetched doc: id=%s, activityId=%s, name=%s", doc.ID, doc.ActivityID, doc.Name)
		log.Printf("[consumer] Indexing session %s in Solr", ev.SessionID)
		log.Printf("[consumer] Document to index: id=%s, activityId=%s, name=%s, sport=%s, site=%s, start_dt=%s", 
			doc.ID, doc.ActivityID, doc.Name, doc.Sport, doc.Site, doc.StartAt)
		if err := c.repo.Upsert(ctx, *doc); err != nil {
			log.Printf("[consumer] ERROR: solr upsert error for session %s: %v", ev.SessionID, err)
			return
		}
		log.Printf("[consumer] SUCCESS: indexed session %s (activityId=%s, name=%s)", ev.SessionID, doc.ActivityID, doc.Name)
	}
}

func (c *Consumer) fetchDoc(sessionID string) (*domain.SearchDoc, error) {
	url := fmt.Sprintf("%s/sessions/%s/search-doc", c.activity, sessionID)
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
