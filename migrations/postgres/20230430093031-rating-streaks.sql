
-- +migrate Up
create table rating_streaks (
    id serial primary key,
    name varchar(200) not null,
    start_date timestamp with time zone not null,
    end_date timestamp with time zone not null,
    unique (name,start_date,end_date)
);
-- +migrate Down

drop table rating_streaks;
