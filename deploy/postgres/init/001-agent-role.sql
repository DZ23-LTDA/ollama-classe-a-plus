-- Local development role grants for the agent runtime.
-- The role and password are created by POSTGRES_USER/POSTGRES_PASSWORD in Compose.
-- Keep the application role non-superuser so FORCE ROW LEVEL SECURITY is effective.
\connect ollama_agent
GRANT USAGE, CREATE ON SCHEMA public TO ollama_agent;
