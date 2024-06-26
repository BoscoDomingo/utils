CREATE FUNCTION PUBLIC.ROUND(real, int)
RETURNS REAL AS
$f$
	SELECT ROUND($1::numeric, $2)::REAL;
$f$ language SQL IMMUTABLE;


CREATE FUNCTION PUBLIC.ROUND(float, int)
RETURNS FLOAT AS
$f$
	SELECT ROUND($1::numeric, $2)::FLOAT;
$f$ language SQL IMMUTABLE;