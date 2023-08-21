-- DAWGS Property Graph Partitioned Layout for PostgreSQL

-- Notes on TOAST:
--
-- Graph entity properties are stored in a JSONB column at the end of the row. There is a soft-limit of 2KiB for rows in
-- a PostgreSQL database page. The database will compress this value in an attempt not to exceed this limit. Once a
-- compressed value reaches the absolute limit of what the database can do to either compact it or give it more of the
-- 8 KiB page size limit, the database evicts the value to an associated TOAST (The Oversized-Attribute Storage Technique)
-- table and creates a reference to the entry to be joined upon fetch of the row.
--
-- TOAST comes with certain performance caveats that can affect access time anywhere from a factor 3 to 6 times. It is
-- in the best interest of the database user that the properties of a graph entity never exceed this limit in large
-- graphs.

-- We need the tri-gram extension to create a GIN text-search index. The goal here isn't full-text search, in which
-- case ts_vector and its ilk would be more suited. This particular selection was made to support accelerated lookups
-- for "contains", "starts with" and, "ends with" comparison operations.
create extension if not exists pg_trgm;

-- Table definitions

-- The graph table contains name to ID mappings for graphs contained within the database. Each graph ID should have
-- corresponding table partitions for the node and edge tables.
create table if not exists graph
(
  id   serial,
  name varchar(256) not null,

  primary key (id),
  unique (name)
);

-- The kind table contains name to ID mappings for graph kinds. Storage of these types is necessary to maintain search
-- capability of a database without the origin application that generated it.
create table if not exists kind
(
  id   smallserial,
  name varchar(256) not null,

  primary key (id),
  unique (name)
);

-- The node table is a partitioned table view that partitions over the graph ID that each node belongs to. Nodes may
-- contain a disjunction of up to 8 kinds for creating clique subsets without requiring edges.
create table if not exists node
(
  id         serial,
  graph_id   integer     not null,
  kind_ids   smallint[8] not null,
  properties jsonb       not null,

  primary key (id, graph_id),

  foreign key (graph_id) references graph (id) on delete cascade
) partition by list (graph_id);

-- The storage strategy chosen for the properties JSONB column informs the database of the user's preference to resort
-- to creating a TOAST table entry only after there is no other possible way to inline the row attribute in the current
-- page.
alter table node
  alter column properties set storage main;

-- Index on the graph ID of each node.
-- TODO: Validate if this is needed at all given that graph_id is the partition key. The partition key is composite so I'm unsure.
create index if not exists node_graph_id_index on node using btree (graph_id);

-- Index node kind IDs so that lookups by kind is accelerated.
create index if not exists node_kind_ids_index on node using gin (kind_ids);

-- The edge table is a partitioned table view that partitions over the graph ID that each edge belongs to.
create table if not exists edge
(
  id         serial,
  graph_id   integer  not null references graph (id),
  start_id   integer  not null,
  end_id     integer  not null,
  kind_id    smallint not null,
  properties jsonb    not null,

  primary key (id, graph_id),

  foreign key (graph_id) references graph (id) on delete cascade
) partition by list (graph_id);

-- The storage strategy chosen for the properties JSONB column informs the database of the user's preference to resort
-- to creating a TOAST table entry only after there is no other possible way to inline the row attribute in the current
-- page.
alter table edge
  alter column properties set storage main;


-- Index on the graph ID of each edge.
-- TODO: Validate if this is needed at all given that graph_id is the partition key. The partition key is composite so I'm unsure.
create index if not exists edge_graph_id_index on edge using btree (graph_id);

-- Index on the start vertex of each edge.
create index if not exists edge_start_id_index on edge using btree (start_id);

-- Index on the start vertex of each edge.
create index if not exists edge_end_id_index on edge using btree (end_id);

-- Index on the kind of each edge.
create index if not exists edge_kind_index on edge using btree (kind_id);
