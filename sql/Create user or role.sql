BEGIN;
	CREATE USER <username> WITH PASSWORD ''; -- If already MD5 or SCRAM-SHA-256 hashed then no extra encryption will be performed
	-- DROP USER IF EXISTS <username>;

	GRANT USAGE ON SCHEMA <schema> TO <username>;
	-- REVOKE ALL ON SCHEMA <schema> FROM <username>;

	GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA <schema> TO <username>;
	-- REVOKE ALL ON ALL TABLES IN SCHEMA <schema> FROM <username>;

	ALTER DEFAULT PRIVILEGES IN SCHEMA <schema> GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO <username>;
	-- ALTER DEFAULT PRIVILEGES IN SCHEMA <schema> REVOKE ALL ON TABLES FROM <username>;
COMMIT;

-- Change user password
BEGIN;
	ALTER USER <username> WITH PASSWORD '<hashed_password>';
COMMIT;

-- Double check (PostgreSQL-specific)
SELECT * FROM pg_authid;