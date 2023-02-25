
-- +migrate Up

create table if not exists merged_dishes (
    id serial primary key,
    name varchar(1000) not null,
    location_id int not null,
    created_at timestamp with time zone not null,
    updated_at timestamp with time zone,
    unique (name,location_id),
    constraint fk_merged_dishes_location_id foreign key (location_id) references locations(id) on delete restrict
);

alter table dishes
    add column merged_dish_id int default null,
    add column merged_at timestamp with time zone default null,
    add constraint fk_dishes_merged_dish_id foreign key (merged_dish_id)
        references merged_dishes(id) on delete set null;

alter table dish_ratings
    drop constraint dish_ratings_pkey,
    add column id serial primary key;

-- +migrate Down

-- delete all but the most recent rating. Otherwise we cannot revert to the old primary key as it would
-- forbid multiple ratings by the same user
delete from dish_ratings d1 where date <>
    (select max(d2.date) from dish_ratings d2 where d1.user_id = d2.user_id);


alter table dish_ratings
    drop constraint dish_ratings_pkey,
    drop column id,
    add  primary key(dish_id,user_id);

alter table dishes
    drop constraint fk_dishes_merged_dish_id,
    drop column merged_dish_id,
    drop column merged_at;

drop table  merged_dishes;
