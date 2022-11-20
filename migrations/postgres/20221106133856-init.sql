
-- +migrate Up

-- +migrate Up
create table if not exists users (
    id serial primary key,
    email varchar(200) not null unique,
    created timestamp with time zone not null
);
comment on column users.email is 'Email of user. Main selector for user in user facing queries';
comment on column users.created is 'Time when this user was first created';


create table if not exists locations (
    id serial primary key,
    name varchar(200) not null unique,
    created timestamp with time zone not null
);


create table if not exists dishes (
    id serial primary key,
    location_id int not null,
    name varchar(1000) not null,
    constraint fk_dishes_locations_location_id foreign key (location_id) references locations(id) on delete restrict,
    unique (name,location_id)
);


create table if not exists dish_occurrences (
    id serial primary key,
    dish_id int not null,
    date timestamp with time zone not null,
    constraint fk_dish_occurrences_dish_id foreign key (dish_id) references dishes(id)  on delete cascade
);

create table if not exists dish_ratings (
    dish_id int not null,
    user_id int not null,
    date timestamp with time zone not null,
    rating int not null,
    primary key (dish_id,user_id),
    constraint  fk_dish_ratings_dish_id foreign key (dish_id) references dishes(id) on delete cascade,
    constraint fk_dish_ratings_user_id foreign key (user_id) references users(id) on delete cascade
);

-- +migrate Down
drop table dish_ratings;

drop table dish_occurrences;

drop table dishes;

drop table locations;

drop table users;
