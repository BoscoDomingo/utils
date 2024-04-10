-- PostgreSQL-specific
SAVEPOINT before_replication_setup;
BEGIN;

	-- Database config checks (will depend on the replication service you use)
	-- SHOW max_replication_slots;
	-- -- SET max_replication_slots TO 10;
	-- SHOW wal_sender_timeout;
	-- -- ALTER SYSTEM SET wal_sender_timeout TO 0;
	-- SHOW max_wal_senders; -- >= 2x the number of replication slots
	-- -- ALTER SYSTEM SET max_wal_senders TO 20;

	GRANT USAGE ON SCHEMA <schema> TO <user>;
	GRANT SELECT ON ALL TABLES IN SCHEMA <schema> TO <user>;
	-- REVOKE SELECT ON ALL TABLES IN SCHEMA <schema> FROM <user>;
	ALTER DEFAULT PRIVILEGES IN SCHEMA <schema> GRANT SELECT ON TABLES TO <user>;
	-- ALTER DEFAULT PRIVILEGES IN SCHEMA <schema> REVOKE SELECT ON TABLES FROM <user>;
	ALTER ROLE <user> WITH REPLICATION;


	-- CREATE PUBLICATION <publication_name> FOR TABLE table2, table4, table8;
	CREATE PUBLICATION <publication_name> FOR TABLES IN SCHEMA <schema>;
	-- DROP PUBLICATION <publication_name>;

	SELECT pg_create_logical_replication_slot('replication_slot_name', 'pgoutput'); -- can also be 'wal2json'
	-- SELECT pg_drop_replication_slot('replication_slot_name');

COMMIT;
-- ROLLBACK TO SAVEPOINT before__replication_setup;