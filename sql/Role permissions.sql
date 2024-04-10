--
-- THIS LIST IS NOT EXHAUSTIVE. PLEASE CHECK ORIGINAL DOCUMENTATION FOR YOUR DATABASE
--

-- ACCESS TO ALL SCHEMAS AND TABLES (PostgreSQL-specific)
GRANT pg_read_all_data TO <user>;

--ACCESS DB
GRANT  CONNECT ON DATABASE <database>  TO <username>;
REVOKE CONNECT ON DATABASE <database> FROM <username>;

--ACCESS SCHEMA
GRANT  USAGE   ON SCHEMA <schema>  TO <username>;
REVOKE ALL     ON SCHEMA <schema> FROM <username>;

--ACCESS TABLES
GRANT SELECT                         ON ALL TABLES IN SCHEMA <schema> TO read_only;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA <schema> TO read_write;
GRANT ALL                            ON ALL TABLES IN SCHEMA <schema> TO admin;
REVOKE ALL ON ALL TABLES IN SCHEMA <schema> FROM <username>;

-- FUTURE TABLES
ALTER DEFAULT PRIVILEGES IN SCHEMA <schema> GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO <username>;
ALTER DEFAULT PRIVILEGES IN SCHEMA <schema> REVOKE ALL ON TABLES FROM <username>;