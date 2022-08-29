package dishRepo

const createUserTable = `create table if not exists users (
    id int auto_increment primary key comment 'Primary Key',
    email varchar(200) not null unique  comment 'Email of user. Main selector for user in user facing queries',
    created datetime not null comment 'Time when this user was first created'
) comment 'Store all known users' ;`

const createDishTable = `create table if not exists dishes (
    id int auto_increment primary key comment 'Primary Key',
    name varchar(1000) not null unique comment 'Name of the dish'
) comment 'Store all known dishes';`

const createDishOccurrencesTable = `create table if not exists dish_occurrences (
    id int auto_increment primary key comment 'Primary Key',
    dish_id int not null comment 'ID of the dish this occurrence is about',
    date datetime not null comment 'Date when the dish was served',
	constraint fk_dish_occurrences_dish_id foreign key (dish_id) references dishes(id)  on delete cascade 
) comment 'Store each day when a dish was served' ;`

const createDishRatingsTable = `create table if not exists dish_ratings (
   dish_id int not null comment 'Rated dish',
   user_id int not null comment 'User who rated/voted',
   date datetime not null comment 'Time of rating',
   rating int not null comment 'Rating given by the user',
   constraint  primary key (dish_id,user_id),
   constraint  fk_dish_ratings_dish_id foreign key (dish_id) references dishes(id) on delete cascade,
   constraint fk_dish_ratings_user_id foreign key (user_id) references users(id) on delete cascade
) comment 'Stores a user\'s rating for a given dish';`
