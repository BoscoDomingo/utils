-- PostgreSQL-specific
select pg_column_size(a) AS real_size
from (values
	(99.99::real),
	(0.01::real),
	(10.0::real),
	(21::real),
	(99.99::real))
s(a);

select pg_column_size(a) as numeric_size
from (values
	(99.99::numeric(4,2)),
	(0.01::numeric(4,2)),
	(10.0::numeric(4,2)),
	(21::numeric(4,2)),
	(99.99::numeric(4,2)))
s(a);