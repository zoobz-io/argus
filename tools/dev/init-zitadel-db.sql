-- Create Zitadel database and user.
-- Runs once on first postgres startup via docker-entrypoint-initdb.d.
CREATE USER zitadel WITH PASSWORD 'zitadel';
CREATE DATABASE zitadel OWNER zitadel;
