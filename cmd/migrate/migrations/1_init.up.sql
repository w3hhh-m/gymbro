CREATE TABLE IF NOT EXISTS Exercises
(
    exercise_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    picture TEXT
);

CREATE TABLE IF NOT EXISTS MuscleGroups
(
    muscle_group_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS ExerciseMuscleGroups
(
    exercise_id INT NOT NULL,
    muscle_group_id INT NOT NULL,
    PRIMARY KEY (exercise_id, muscle_group_id),
    FOREIGN KEY (exercise_id) REFERENCES Exercises (exercise_id) ON DELETE CASCADE,
    FOREIGN KEY (muscle_group_id) REFERENCES MuscleGroups (muscle_group_id) ON DELETE CASCADE
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
    points INT NOT NULL DEFAULT 0,
    date_of_birth date,
    google_id VARCHAR(100) UNIQUE,
    fk_clan_id TEXT REFERENCES Clans(clan_id) ON DELETE SET NULL,
    fk_gym_id INT REFERENCES Gyms(gym_id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT false,
    last_active TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS UserExerciseMaxWeights
(
    user_id TEXT NOT NULL,
    exercise_id INT NOT NULL,
    max_weight INT NOT NULL DEFAULT 0,
    reps INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, exercise_id),
    FOREIGN KEY (user_id) REFERENCES Users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (exercise_id) REFERENCES Exercises(exercise_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS Workouts
(
    workout_id TEXT PRIMARY KEY,
    fk_user_id TEXT REFERENCES Users(user_id) ON DELETE CASCADE,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    points INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS Records
(
    record_id TEXT PRIMARY KEY,
    fk_workout_id  TEXT REFERENCES Workouts(workout_id) ON DELETE CASCADE,
    fk_exercise_id integer REFERENCES Exercises(exercise_id) ON DELETE CASCADE,
    reps INT NOT NULL,
    weight INT NOT NULL,
    points INT NOT NULL
);

CREATE TABLE IF NOT EXISTS Subscriptions (
    subscription_id TEXT PRIMARY KEY,
    fk_user_id TEXT REFERENCES Users(user_id),
    fk_gym_id INT REFERENCES Gyms(gym_id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO Users (user_id, username, email)
VALUES ('SECRET_USER', 'SECRET', 'SECRET@GYMBRO.COM');

ALTER TABLE Clans
ADD CONSTRAINT fk_owner_id
FOREIGN KEY (fk_owner_id) REFERENCES Users(user_id);