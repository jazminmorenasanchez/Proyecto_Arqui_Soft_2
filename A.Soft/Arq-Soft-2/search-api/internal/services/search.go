package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/sporthub/search-api/internal/domain"
)

type solrRepo interface {
	Search(ctx context.Context, q, sport, site, date, sort string, page, size int) (*domain.Result, error)
}

type localCache interface {
	Get(key string) interface{}
	Set(key string, val interface{}, ttl time.Duration)
	Delete(key string)
}
type distCache interface {
	Get(key string, out any) bool
	Set(key string, val any, ttl time.Duration)
	Delete(key string)
}

type Service struct {
	repo solrRepo
	lc   localCache
	dc   distCache
	ttl  time.Duration
}

func NewSearchService(r solrRepo, lc localCache, dc distCache, ttl time.Duration) *Service {
	return &Service{repo: r, lc: lc, dc: dc, ttl: ttl}
}

func (s *Service) Search(ctx context.Context, q, sport, site, date, sort string, page, size int) (*domain.Result, error) {
	key := s.key(q, sport, site, date, sort, page, size)

	// 1) local cache
	if v := s.lc.Get(key); v != nil {
		if res, ok := v.(*domain.Result); ok {
			return res, nil
		}
	}
	// 2) distributed cache
	var cached domain.Result
	if s.dc.Get(key, &cached) {
		s.lc.Set(key, &cached, s.ttl)
		return &cached, nil
	}

	// 3) Solr
	res, err := s.repo.Search(ctx, q, sport, site, date, sort, page, size)
	if err != nil {
		return nil, err
	}

	// 4) save caches
	s.lc.Set(key, res, s.ttl)
	s.dc.Set(key, res, s.ttl)
	return res, nil
}

func (s *Service) Bust(key string) {
	s.lc.Delete(key)
	s.dc.Delete(key)
}

func (s *Service) key(q, sport, site, date, sort string, page, size int) string {
	raw := fmt.Sprintf("%s|%s|%s|%s|%s|%d|%d", q, sport, site, date, sort, page, size)
	h := sha1.Sum([]byte(raw))
	return "q:" + hex.EncodeToString(h[:])
}
