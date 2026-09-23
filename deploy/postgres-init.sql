-- The official postgres image creates POSTGRES_USER (ollama_agent) as a
-- SUPERUSER, and PostgreSQL always bypasses Row Level Security for superusers —
-- even with FORCE ROW LEVEL SECURITY. Connecting the application as that
-- superuser silently voids the per-tenant RLS policies (one tenant could read
-- every other tenant's missions and events).
--
-- Provision a dedicated, non-superuser application role. The agent runtime must
-- connect as this role (see OLLAMA_AGENT_DATABASE_URL / the test DSN): its
-- migrations create the tables, so it owns them, and FORCE ROW LEVEL SECURITY
-- then actually applies to it. NOBYPASSRLS is explicit belt-and-braces.
CREATE ROLE ollama_app WITH LOGIN PASSWORD 'change-me-local-only' NOSUPERUSER NOBYPASSRLS NOCREATEROLE NOCREATEDB;
GRANT CONNECT ON DATABASE ollama_agent TO ollama_app;
GRANT USAGE, CREATE ON SCHEMA public TO ollama_app;
