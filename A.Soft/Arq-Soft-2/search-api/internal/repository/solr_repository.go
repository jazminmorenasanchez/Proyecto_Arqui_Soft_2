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
	
	// Usar edismax query parser para búsquedas más flexibles y mejor relevancia
	params.Set("defType", "edismax")
	
	// Con edismax, usamos las palabras simples y dejamos que edismax haga el matching
	// El analyzer de Solr (LowerCaseFilterFactory) hace que la búsqueda sea case-insensitive
	query := "*:*"
	var cleanWords []string
	if q != "" {
		// Limpiar y normalizar la query (convertir a minúsculas para consistencia)
		// Aunque el analyzer lo hace, es bueno normalizar aquí también
		q = strings.TrimSpace(strings.ToLower(q))
		if q != "" {
			// Dividir la query en palabras
			words := strings.Fields(q)
			for _, word := range words {
				// Escapar solo caracteres especiales que puedan romper la query
				escaped := escapeForSolrQuery(word)
				if escaped != "" {
					cleanWords = append(cleanWords, escaped)
				}
			}
			if len(cleanWords) > 0 {
				// Con edismax, pasamos las palabras separadas por espacios
				// edismax buscará en los campos especificados en qf usando el analyzer
				query = strings.Join(cleanWords, " ")
			}
		}
	}
	params.Set("q", query)
	
	// Configurar campos de búsqueda para edismax (query fields)
	// edismax buscará en estos campos cuando se use el parámetro q
	// El analyzer text_general con LowerCaseFilterFactory hace que la búsqueda sea case-insensitive
	// Usamos name_txt con mayor relevancia (^2) para que coincidencias en el nombre tengan más peso
	params.Set("qf", "name_txt^2 sport_s^1 site_s^1 instructor_s^1")
	
	// Configurar campos de frase para boosting de coincidencias exactas
	// Esto da más relevancia a frases completas que coincidan
	params.Set("pf", "name_txt^3")
	
	// Minimum match dinámico según el número de palabras:
	// - 1 palabra: mm=1 (búsqueda parcial flexible - permite encontrar "futbol" en "futbol 5" y "futbol 7")
	// - 2+ palabras: mm=100% (búsqueda exacta - todas las palabras deben coincidir)
	// Esto asegura que búsquedas exactas como "Futbol 5 Gran 7" solo devuelvan resultados que contengan todas las palabras
	if len(cleanWords) == 1 {
		// Búsqueda parcial: al menos una palabra debe coincidir
		params.Set("mm", "1")
	} else if len(cleanWords) > 1 {
		// Búsqueda exacta: todas las palabras deben coincidir
		params.Set("mm", "100%")
	} else {
		// Sin palabras (query vacía): mantener comportamiento por defecto
		params.Set("mm", "1")
	}
	
	// Usar el operador OR por defecto para que cualquier palabra coincida
	// Esto es importante para búsquedas parciales
	params.Set("q.op", "OR")

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
	log.Printf("[solr] Search query: %s", u)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	resp, err := r.http.Do(req)
	if err != nil {
		log.Printf("[solr] Search error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	log.Printf("[solr] Search response status: %d, body length: %d", resp.StatusCode, len(b))

	var sr solrResponse
	if err := json.Unmarshal(b, &sr); err != nil {
		return nil, err
	}

	out := &domain.Result{Total: sr.Response.NumFound, Page: page, Size: size}
	log.Printf("[solr] Found %d documents, returning %d", sr.Response.NumFound, len(sr.Response.Docs))
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
		log.Printf("[solr] Found doc: id=%s, activity_id=%s, name=%s", doc.ID, doc.ActivityID, doc.Name)
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
			"name_txt":      doc.Name,
			"sport_s":       doc.Sport,
			"site_s":        doc.Site,
			"instructor_s":  doc.Instructor,
			"difficulty_i":  doc.Difficulty,
			"price_f":      doc.Price,
			"tags_ss":       doc.Tags,
			"updated_dt":    doc.UpdatedAt,
		}
		// No incluimos session_id: no indexamos sesiones, solo actividades
		// Solo incluir campos de fecha si tienen valores válidos (no vacíos)
		// Solr rechaza strings vacíos para campos de tipo pdate
		if doc.StartAt != "" {
			solrDoc["start_dt"] = doc.StartAt
		}
		if doc.EndAt != "" {
			solrDoc["end_dt"] = doc.EndAt
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

// escapeForSolr escapa caracteres especiales de Solr (para uso con wildcards)
func escapeForSolr(s string) string {
	// Escapar caracteres especiales de Solr: + - && || ! ( ) { } [ ] ^ " ~ * ? : \
	specialChars := []string{"+", "-", "&&", "||", "!", "(", ")", "{", "}", "[", "]", "^", "\"", "~", "*", "?", ":", "\\"}
	result := s
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}

// escapeForSolrQuery escapa caracteres especiales para queries con edismax
// NO escapa * porque edismax no usa wildcards en el campo q
func escapeForSolrQuery(s string) string {
	// Escapar solo caracteres que pueden romper la query de edismax
	// NO escapamos * porque no lo usamos con edismax
	specialChars := []string{"+", "-", "&&", "||", "!", "(", ")", "{", "}", "[", "]", "^", "\"", "~", "?", ":", "\\"}
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
