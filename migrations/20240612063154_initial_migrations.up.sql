create table users (
    id integer primary key,
    username text not null,
    password text not null
);

create table rooms (
    id integer primary key,
    room_name text not null
);

create table room_participants (
    id integer primary key,
    room_id integer not null,
    user_id integer not null
);

create table chat_histories (
    id integer primary key, 
    room_id integer not null,
    sender_id integer not null,
    recipient_id integer not null,
    message text not null
);