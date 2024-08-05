CREATE TABLE IF NOT EXISTS Exercises
(
    exercise_id serial PRIMARY KEY,
    name varchar(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS Gyms (
    gym_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    address TEXT,
    description TEXT
);

INSERT INTO Gyms (gym_id, name, address, description)
VALUES (0, 'Default', 'No address', 'Default gym entry for users without a specific gym');

CREATE TABLE IF NOT EXISTS Clans (
    clan_id TEXT PRIMARY KEY,
    fk_owner_id TEXT,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    points INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO Clans (clan_id, fk_owner_id, name, description, points)
VALUES ('0', 'SECRET_USER', 'Default', 'Default clan for users without a specific clan', 0);

CREATE TABLE IF NOT EXISTS Users
(
    user_id TEXT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash TEXT,
    date_of_birth date,
    google_id VARCHAR(100) UNIQUE,
    fk_clan_id TEXT REFERENCES Clans(clan_id),
    fk_gym_id integer REFERENCES Gyms(gym_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO Users (user_id, username, email)
VALUES ('SECRET_USER', 'SECRET', 'SECRET@GYMBRO.COM');

CREATE TABLE IF NOT EXISTS Workouts
(
    workout_id TEXT PRIMARY KEY,
    fk_user_id TEXT REFERENCES Users(user_id),
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    points INT NOT NULL DEFAULT 0,
    is_active BOOLEAN
);

CREATE TABLE IF NOT EXISTS Records
(
    record_id TEXT PRIMARY KEY,
    fk_workout_id  TEXT REFERENCES Workouts(workout_id) NOT NULL,
    fk_exercise_id integer REFERENCES Exercises(exercise_id) NOT NULL,
    reps integer NOT NULL,
    weight integer NOT NULL
);

CREATE TABLE IF NOT EXISTS Subscriptions (
    subscription_id TEXT PRIMARY KEY,
    fk_user_id TEXT REFERENCES Users(user_id),
    fk_gym_id integer REFERENCES Gyms(gym_id),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE Clans
ADD CONSTRAINT fk_owner_id
FOREIGN KEY (fk_owner_id) REFERENCES Users(user_id);