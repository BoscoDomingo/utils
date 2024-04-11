-- List schema privileges for current user
WITH "names"("name") AS (
  SELECT n.nspname AS "name"
    FROM pg_catalog.pg_namespace n
      WHERE n.nspname !~ '^pg_'
        AND n.nspname <> 'information_schema'
) SELECT "name",
  pg_catalog.has_schema_privilege(current_user, "name", 'CREATE') AS "create",
  pg_catalog.has_schema_privilege(current_user, "name", 'USAGE') AS "usage"
    FROM "names";

-- List privileges by username
SELECT
  n.nspname as "Schema",
  CASE c.relkind
    WHEN 'r' THEN 'table'
    WHEN 'v' THEN 'view'
    WHEN 'm' THEN 'materialized view'
    WHEN 'S' THEN 'sequence'
    WHEN 'f' THEN 'foreign table'
  END as "Type",
  c.relname as "Name"
FROM pg_catalog.pg_class c
LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE pg_catalog.array_to_string(c.relacl, E'\n') LIKE '%username%';