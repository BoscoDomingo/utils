SAVEPOINT before_index_drop;
BEGIN;
	DROP INDEX IF EXISTS "schema"."table" RESTRICT;
COMMIT;

-- ROLLBACK TO SAVEPOINT before_index_drop;