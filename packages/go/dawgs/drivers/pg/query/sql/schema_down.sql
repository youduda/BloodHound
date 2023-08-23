-- Drop all tables in order of dependency.
drop table if exists node;
drop table if exists edge;
drop table if exists kind;
drop table if exists graph;

-- Pull the tri-gram extension.
drop extension if exists pg_trgm;
