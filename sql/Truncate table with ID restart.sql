SAVEPOINT before_truncate;
BEGIN;
	TRUNCATE TABLE <TABLE> RESTART IDENTITY;
COMMIT;

-- ROLLBACK TO SAVEPOINT before_truncate;