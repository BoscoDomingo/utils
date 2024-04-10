-- PostgreSQL-specific

-- See all published tables and columns
SELECT * FROM pg_publication_tables ORDER BY pubname, schemaname, tablename;

-- See all replication slots
SELECT * FROM pg_catalog.pg_replication_slots;

-- See current processes using replications
SELECT * FROM pg_stat_replication;

-- See remaining changes from a replication slot (won't work for active slots, e.g. the ones used by processes in the query above)
SELECT COUNT(*) FROM pg_logical_slot_peek_changes('replication_slot_name', null, null);
SELECT COUNT(*) FROM pg_logical_slot_peek_binary_changes('replication_slot_name', null, null, 'proto_version', '1', 'publication_names', '<publication_name>');