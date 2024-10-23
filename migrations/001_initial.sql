-- Write your migrate up statements here
create extension if not exists "uuid-ossp";

-- function to update timestamps if a table has an updated_at column
create or replace function set_updated_at()
  returns trigger
  as $$
begin
  new.updated_at = now();
  return NEW;
end;
$$
language plpgsql;

-- automate trigger creation
create or replace function trigger_updated_at(tablename regclass)
  returns void
  as $$
begin
  execute format('create trigger set_updated_at
        before update
        on %s
        for each row
        when (OLD is distinct from NEW)
    execute function set_updated_at();', tablename);
end;
$$
language plpgsql;

create table if not exists users (
id uuid not null default uuid_generate_v4() primary key,
first_name text not null,
  last_name text not null,
  email text not null,
  created_at timestamptz not null default now(),
  updated_at timestamptz,
  unique(email)
);

select
  trigger_updated_at('users');

create table if not exists clubs (
id uuid not null default uuid_generate_v4() primary key,
  name text not null,
  president_id uuid not null references users(id) on delete cascade,

  created_at timestamptz not null default now(),
  updated_at timestamptz,
  unique(name)
  );

select
  trigger_updated_at('clubs');


---- create above / drop below ----

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
drop table if exists clubs cascade;
drop table if exists users cascade;
drop extension if exists "uuid-ossp";