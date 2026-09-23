package grok

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Evidence struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	Title       string    `json:"title,omitempty"`
	Excerpt     string    `json:"excerpt,omitempty"`
	Kind        string    `json:"kind"` // live_web, live_x, memory
	Publisher   string    `json:"publisher,omitempty"`
	RetrievedAt time.Time `json:"retrieved_at"`
	SHA256      string    `json:"sha256,omitempty"`
}

type LiveQuery struct {
	Query        string     `json:"query"`
	Web          bool       `json:"web"`
	X            bool       `json:"x"`
	Memory       []Evidence `json:"memory,omitempty"`
	Live         []Evidence `json:"live,omitempty"`
	MaxEvidence  int        `json:"max_evidence,omitempty"`
	RequireHTTPS bool       `json:"require_https"`
}

type LiveReport struct {
	Query          string     `json:"query"`
	Evidence       []Evidence `json:"evidence"`
	Citations      []Citation `json:"citations"`
	Confirmed      []string   `json:"confirmed,omitempty"`
	Unconfirmed    []string   `json:"unconfirmed,omitempty"`
	Conflicts      []string   `json:"conflicts,omitempty"`
	Confidence     float64    `json:"confidence"`
	MemoryEvidence int        `json:"memory_evidence"`
	LiveEvidence   int        `json:"live_evidence"`
	GeneratedAt    time.Time  `json:"generated_at"`
}

type LiveMode struct{}

func NewLiveMode() *LiveMode { return &LiveMode{} }

func (LiveMode) Build(query LiveQuery) (LiveReport, error) {
	query.Query = strings.TrimSpace(query.Query)
	if query.Query == "" {
		return LiveReport{}, errors.New("live query is required")
	}
	if query.MaxEvidence <= 0 || query.MaxEvidence > 64 {
		query.MaxEvidence = 32
	}
	evidence := make([]Evidence, 0, len(query.Memory)+len(query.Live))
	for _, item := range query.Memory {
		item.Kind = "memory"
		if item.RetrievedAt.IsZero() {
			item.RetrievedAt = time.Now().UTC()
		}
		if item.ID == "" {
			item.ID = evidenceID(item)
		}
		evidence = append(evidence, item)
	}
	for _, item := range query.Live {
		if item.Kind == "" {
			item.Kind = "live_web"
		}
		if item.RetrievedAt.IsZero() {
			item.RetrievedAt = time.Now().UTC()
		}
		if item.ID == "" {
			item.ID = evidenceID(item)
		}
		evidence = append(evidence, item)
	}
	for _, item := range evidence {
		if query.RequireHTTPS && item.URL != "" {
			parsed, err := url.Parse(item.URL)
			if err != nil || parsed.Scheme != "https" {
				return LiveReport{}, errors.New("live evidence URL must use HTTPS")
			}
		}
	}
	if len(evidence) > query.MaxEvidence {
		evidence = evidence[:query.MaxEvidence]
	}
	sort.SliceStable(evidence, func(i, j int) bool {
		if evidence[i].RetrievedAt.Equal(evidence[j].RetrievedAt) {
			return evidence[i].ID < evidence[j].ID
		}
		return evidence[i].RetrievedAt.Before(evidence[j].RetrievedAt)
	})
	report := LiveReport{Query: query.Query, Evidence: evidence, GeneratedAt: time.Now().UTC()}
	for _, item := range evidence {
		if item.Kind == "memory" {
			report.MemoryEvidence++
		} else {
			report.LiveEvidence++
		}
		if item.URL != "" {
			report.Citations = append(report.Citations, Citation{URL: item.URL, Title: item.Title, Excerpt: item.Excerpt, Publisher: item.Publisher})
		}
	}
	if len(evidence) > 0 {
		report.Confidence = float64(report.LiveEvidence*2+report.MemoryEvidence) / float64(len(evidence)*2)
	}
	if report.LiveEvidence == 0 {
		report.Unconfirmed = append(report.Unconfirmed, "nenhuma fonte ao vivo foi consultada")
	} else if report.LiveEvidence == 1 {
		report.Unconfirmed = append(report.Unconfirmed, "a evidência ao vivo ainda não foi corroborada por outra fonte")
	} else {
		report.Confirmed = append(report.Confirmed, "há pelo menos duas evidências ao vivo disponíveis para comparação")
	}
	return report, nil
}

func evidenceID(item Evidence) string {
	hash := sha256.Sum256([]byte(item.URL + "\x00" + item.Title + "\x00" + item.Excerpt + "\x00" + item.Kind))
	return "ev_" + hex.EncodeToString(hash[:8])
}
