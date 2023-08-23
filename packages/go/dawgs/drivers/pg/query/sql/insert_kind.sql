-- Creates a new kind definition and returns the resulting ID.
insert into kind (name)
values (@name)
returning id;
