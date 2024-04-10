create extension if not exists moddatetime schema extensions;

BEGIN;

-- assuming the table has a timestamp column "updated_at"
-- this trigger will set the "updated_at" column to the current timestamp for every update
create trigger updated_at_trigger before
update
  on <table> for each row execute procedure "extensions"."moddatetime"(updated_at);

COMMIT;