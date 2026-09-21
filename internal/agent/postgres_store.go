package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct {
	db      *sql.DB
	timeout time.Duration
}

func OpenPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("postgres DSN is required")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	store := &PostgresStore{db: db, timeout: 10 * time.Second}
	pingCtx, cancel := context.WithTimeout(ctx, store.timeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := store.Migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *PostgresStore) Migrate(ctx context.Context) error {
	if s == nil || s.db == nil {
		return errors.New("postgres store is not initialized")
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS agent_missions (id TEXT PRIMARY KEY, version BIGINT NOT NULL, objective TEXT NOT NULL, model TEXT NOT NULL DEFAULT '', workspace TEXT NOT NULL DEFAULT '', project_id TEXT NOT NULL DEFAULT '', auto_run BOOLEAN NOT NULL DEFAULT FALSE, state TEXT NOT NULL, plan JSONB NOT NULL, approvals JSONB NOT NULL, artifacts JSONB NOT NULL, last_error TEXT NOT NULL DEFAULT '', created_at TIMESTAMPTZ NOT NULL, updated_at TIMESTAMPTZ NOT NULL, completed_at TIMESTAMPTZ NULL)`,
		`CREATE TABLE IF NOT EXISTS agent_events (id TEXT PRIMARY KEY, mission_id TEXT NOT NULL, type TEXT NOT NULL, step_id TEXT NOT NULL DEFAULT '', payload JSONB NULL, created_at TIMESTAMPTZ NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS agent_events_mission_created_idx ON agent_events (mission_id, created_at, id)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}
func (s *PostgresStore) GetMission(id string) (Mission, error) {
	var mission Mission
	var plan, approvals, artifacts []byte
	var completed sql.NullTime
	err := s.db.QueryRow(`SELECT id,version,objective,model,workspace,project_id,auto_run,state,plan,approvals,artifacts,last_error,created_at,updated_at,completed_at FROM agent_missions WHERE id=$1`, id).Scan(&mission.ID, &mission.Version, &mission.Objective, &mission.Model, &mission.Workspace, &mission.ProjectID, &mission.AutoRun, &mission.State, &plan, &approvals, &artifacts, &mission.LastError, &mission.CreatedAt, &mission.UpdatedAt, &completed)
	if errors.Is(err, sql.ErrNoRows) {
		return Mission{}, os.ErrNotExist
	}
	if err != nil {
		return Mission{}, err
	}
	if err = json.Unmarshal(plan, &mission.Plan); err != nil {
		return Mission{}, err
	}
	if err = json.Unmarshal(approvals, &mission.Approvals); err != nil {
		return Mission{}, err
	}
	if err = json.Unmarshal(artifacts, &mission.Artifacts); err != nil {
		return Mission{}, err
	}
	if completed.Valid {
		mission.CompletedAt = &completed.Time
	}
	return mission, nil
}
func (s *PostgresStore) ListMissions() ([]Mission, error) {
	rows, err := s.db.Query(`SELECT id FROM agent_missions ORDER BY updated_at ASC,id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var missions []Mission
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		mission, err := s.GetMission(id)
		if err != nil {
			return nil, err
		}
		missions = append(missions, mission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(missions, func(i, j int) bool { return missions[i].UpdatedAt.Before(missions[j].UpdatedAt) })
	return missions, nil
}
func (s *PostgresStore) PutMission(mission Mission) error {
	if strings.TrimSpace(mission.ID) == "" {
		return errors.New("mission id is required")
	}
	plan, _ := json.Marshal(mission.Plan)
	approvals, _ := json.Marshal(mission.Approvals)
	artifacts, _ := json.Marshal(mission.Artifacts)
	_, err := s.db.Exec(`INSERT INTO agent_missions (id,version,objective,model,workspace,project_id,auto_run,state,plan,approvals,artifacts,last_error,created_at,updated_at,completed_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) ON CONFLICT (id) DO UPDATE SET version=EXCLUDED.version,objective=EXCLUDED.objective,model=EXCLUDED.model,workspace=EXCLUDED.workspace,project_id=EXCLUDED.project_id,auto_run=EXCLUDED.auto_run,state=EXCLUDED.state,plan=EXCLUDED.plan,approvals=EXCLUDED.approvals,artifacts=EXCLUDED.artifacts,last_error=EXCLUDED.last_error,updated_at=EXCLUDED.updated_at,completed_at=EXCLUDED.completed_at`, mission.ID, mission.Version, mission.Objective, mission.Model, mission.Workspace, mission.ProjectID, mission.AutoRun, mission.State, plan, approvals, artifacts, mission.LastError, mission.CreatedAt, mission.UpdatedAt, mission.CompletedAt)
	return err
}
func (s *PostgresStore) AppendEvent(event Event) error {
	if event.ID == "" || event.MissionID == "" {
		return errors.New("event id and mission id are required")
	}
	payload, _ := json.Marshal(event.Payload)
	_, err := s.db.Exec(`INSERT INTO agent_events (id,mission_id,type,step_id,payload,created_at) VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT (id) DO NOTHING`, event.ID, event.MissionID, event.Type, event.StepID, payload, event.CreatedAt)
	return err
}
func (s *PostgresStore) ListEvents(missionID string) ([]Event, error) {
	rows, err := s.db.Query(`SELECT id,mission_id,type,step_id,payload,created_at FROM agent_events WHERE mission_id=$1 ORDER BY created_at ASC,id ASC`, missionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []Event
	for rows.Next() {
		var event Event
		var payload []byte
		if err := rows.Scan(&event.ID, &event.MissionID, &event.Type, &event.StepID, &payload, &event.CreatedAt); err != nil {
			return nil, err
		}
		if len(payload) > 0 && string(payload) != "null" {
			if err := json.Unmarshal(payload, &event.Payload); err != nil {
				return nil, err
			}
		}
		events = append(events, event)
	}
	return events, rows.Err()
}
func (s *PostgresStore) SetTimeout(timeout time.Duration) {
	if timeout > 0 {
		s.timeout = timeout
	}
}
func (s *PostgresStore) String() string { return fmt.Sprintf("PostgresStore(timeout=%s)", s.timeout) }
