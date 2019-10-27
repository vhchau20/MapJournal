create table entry_geotag_points (
	ogc_fid serial not null,
	user_name varchar (40),
	entry_name varchar (100),
	asset_name varchar (100)
);

select AddGeometryColumn('entry_geotag_points', 'wkb_geometry', 3857, 'POINT', 2);

grant select, insert on all tables in schema public to tegola;
grant usage, select on all sequences in schema public to tegola;
