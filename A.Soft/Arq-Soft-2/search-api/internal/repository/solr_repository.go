package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/sporthub/search-api/internal/domain"
)

type SolrRepo struct {
	base string
	http *http.Client
}

func NewSolrRepo(solrURL string) *SolrRepo {
	return &SolrRepo{base: strings.TrimRight(solrURL, "/"), http: &http.Client{}}
}

type solrResponse struct {
	Response struct {
		NumFound int                      `json:"numFound"`
		Docs     []map[string]interface{} `json:"docs"`
	} `json:"response"`
}

func (r *SolrRepo) Search(ctx context.Context, q, sport, site, date, sort string, page, size int) (*domain.Result, error) {
	params := url.Values{}
	query := "*:*"
	if q != "" {
		// Búsqueda parcial usando wildcards para permitir coincidencias parciales
		// Para búsquedas con múltiples palabras, usamos AND para que todas las palabras coincidan
		// Escapamos caracteres especiales pero preservamos espacios para búsqueda de múltiples palabras
		words := strings.Fields(q) // Divide la query en palabras individuales
		var conditions []string
		for _, word := range words {
			escaped := escapeForSolr(word)
			conditions = append(conditions, fmt.Sprintf("name_txt:*%s*", escaped))
		}
		if len(conditions) > 0 {
			query = strings.Join(conditions, " AND ")
		}
	}
	params.Set("q", query)

	var fqs []string
	if sport != "" {
		fqs = append(fqs, fmt.Sprintf("sport_s:%q", sport))
	}
	if site != "" {
		fqs = append(fqs, fmt.Sprintf("site_s:%q", site))
	}
	if date != "" {
		fqs = append(fqs, fmt.Sprintf("start_dt:[%sT00:00:00Z TO *]", date))
	}
	for _, fq := range fqs {
		params.Add("fq", fq)
	}

	if sort == "" {
		sort = "start_dt asc"
	}
	params.Set("sort", sort)
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	params.Set("start", fmt.Sprintf("%d", start))
	params.Set("rows", fmt.Sprintf("%d", size))
	params.Set("wt", "json")

	u := fmt.Sprintf("%s/select?%s", r.base, params.Encode())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := r.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)

	var sr solrResponse
	if err := json.Unmarshal(b, &sr); err != nil {
		return nil, err
	}

	out := &domain.Result{Total: sr.Response.NumFound, Page: page, Size: size}
	for _, d := range sr.Response.Docs {
		doc := domain.SearchDoc{
			ID:         asString(d["id"]),
			ActivityID: asString(d["activity_id"]),
			SessionID:  asString(d["session_id"]),
			Name:       asString(d["name_txt"]),
			Sport:      asString(d["sport_s"]),
			Site:       asString(d["site_s"]),
			Instructor: asString(d["instructor_s"]),
			StartAt:    asString(d["start_dt"]),
			EndAt:      asString(d["end_dt"]),
			Difficulty: asInt(d["difficulty_i"]),
			Price:      asFloat(d["price_f"]),
			Tags:       asStrings(d["tags_ss"]),
			UpdatedAt:  asString(d["updated_dt"]),
		}
		out.Docs = append(out.Docs, doc)
	}
	return out, nil
}

// (Estos helpers son mínimos; en la Parte 3 agregaremos Upsert/Delete para el consumer)
func (r *SolrRepo) Upsert(ctx context.Context, docs ...domain.SearchDoc) error {
	// Transformar documentos al formato que Solr espera (con sufijos en los campos)
	solrDocs := make([]map[string]any, 0, len(docs))
	for _, doc := range docs {
		solrDoc := map[string]any{
			"id":            doc.ID,
			"activity_id":   doc.ActivityID,
			"session_id":    doc.SessionID,
			"name_txt":      doc.Name,
			"sport_s":       doc.Sport,
			"site_s":        doc.Site,
			"instructor_s":  doc.Instructor,
			"start_dt":      doc.StartAt,
			"end_dt":        doc.EndAt,
			"difficulty_i":  doc.Difficulty,
			"price_f":      doc.Price,
			"tags_ss":       doc.Tags,
			"updated_dt":    doc.UpdatedAt,
		}
		solrDocs = append(solrDocs, solrDoc)
	}
	
	payload := map[string]any{"add": solrDocs, "commit": true}
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	
	u := fmt.Sprintf("%s/update?commit=true", r.base)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := r.http.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	
	// Log de la respuesta para debugging
	log.Printf("[solr] Response status: %d, body: %s", resp.StatusCode, string(body))
	
	// Verificar si Solr retornó un error
	var solrResp map[string]any
	if err := json.Unmarshal(body, &solrResp); err == nil {
		if errObj, ok := solrResp["error"].(map[string]any); ok {
			msg := fmt.Sprintf("%v", errObj["msg"])
			return fmt.Errorf("solr error: %s", msg)
		}
	}
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("solr returned status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}
func (r *SolrRepo) DeleteByID(ctx context.Context, id string) error {
	payload := map[string]any{"delete": map[string]string{"id": id}, "commit": true}
	b, _ := json.Marshal(payload)
	u := fmt.Sprintf("%s/update?commit=true", r.base)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	_, err := r.http.Do(req)
	return err
}

func escape(s string) string { return strings.ReplaceAll(s, " ", "\\ ") }

// escapeForSolr escapa caracteres especiales de Solr y permite búsqueda parcial
func escapeForSolr(s string) string {
	// Escapar caracteres especiales de Solr: + - && || ! ( ) { } [ ] ^ " ~ * ? : \
	specialChars := []string{"+", "-", "&&", "||", "!", "(", ")", "{", "}", "[", "]", "^", "\"", "~", "*", "?", ":", "\\"}
	result := s
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}
func asString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprintf("%v", v)
	}
}
func asInt(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	default:
		return 0
	}
}
func asFloat(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	default:
		return 0
	}
}
func asStrings(v any) []string {
	if v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, x := range arr {
		out = append(out, asString(x))
	}
	return out
}
