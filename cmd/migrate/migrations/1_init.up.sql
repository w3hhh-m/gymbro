CREATE TABLE IF NOT EXISTS users
(
    user_id serial PRIMARY KEY,
    username text NOT NULL UNIQUE,
    email text NOT NULL UNIQUE,
    phone char(11) UNIQUE,
    password text NOT NULL,
    date_of_birth date,
    created_at timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS exercises
(
    exercise_id serial PRIMARY KEY,
    name varchar(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS records
(
    record_id serial PRIMARY KEY,
    fk_user_id  integer REFERENCES users(user_id) NOT NULL,
    fk_exercise_id integer REFERENCES exercises(exercise_id) NOT NULL,
    reps integer NOT NULL,
    weight integer NOT NULL,
    created_at timestamp NOT NULL
);