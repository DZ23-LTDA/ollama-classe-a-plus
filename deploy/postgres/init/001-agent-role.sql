-- Local development role for the agent runtime.
-- Keep the application role non-superuser so FORCE ROW LEVEL SECURITY is effective.
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'ollama_agent') THEN
    CREATE ROLE ollama_agent LOGIN PASSWORD 'change-me-local-only' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
  ELSE
    ALTER ROLE ollama_agent LOGIN PASSWORD 'change-me-local-only' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
  END IF;
END
$$;

SELECT 'CREATE DATABASE ollama_agent OWNER postgres'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'ollama_agent')\gexec

GRANT CONNECT ON DATABASE ollama_agent TO ollama_agent;

\connect ollama_agent
GRANT USAGE, CREATE ON SCHEMA public TO ollama_agent;
