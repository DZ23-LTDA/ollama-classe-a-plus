-- Dedicated NON-superuser application role for the agent runtime.
-- PostgreSQL bypasses Row Level Security for superusers/BYPASSRLS roles even
-- with FORCE ROW LEVEL SECURITY, so the application must NOT connect as the
-- bootstrap superuser. Point OLLAMA_AGENT_DATABASE_URL at this role.
\connect ollama_agent
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'ollama_app') THEN
    CREATE ROLE ollama_app LOGIN PASSWORD 'change-me-local-only' NOSUPERUSER NOBYPASSRLS NOCREATEROLE NOCREATEDB;
  END IF;
END
$$;
GRANT CONNECT ON DATABASE ollama_agent TO ollama_app;
GRANT USAGE, CREATE ON SCHEMA public TO ollama_app;
