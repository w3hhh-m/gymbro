INSERT INTO MuscleGroups (name)
VALUES
    ('Chest'),
    ('Back'),
    ('Legs'),
    ('Shoulders'),
    ('Arms'),
    ('Abs');

INSERT INTO Exercises (name, description, picture)
VALUES
    ('Bench Press', 'A basic chest exercise performed with a barbell or dumbbells.', 'bench_press.png'),
    ('Deadlift', 'A fundamental compound exercise targeting the entire posterior chain.', 'deadlift.png'),
    ('Squat', 'A primary leg exercise that targets the quadriceps and glutes.', 'squat.png'),
    ('Shoulder Press', 'An overhead pressing movement that targets the deltoid muscles.', 'shoulder_press.png'),
    ('Bicep Curl', 'An isolated exercise that targets the biceps.', 'bicep_curl.png'),
    ('Crunch', 'An abdominal exercise focusing on the rectus abdominis.', 'crunch.png');

INSERT INTO ExerciseMuscleGroups (exercise_id, muscle_group_id)
VALUES
    ((SELECT exercise_id FROM Exercises WHERE name = 'Bench Press'), (SELECT muscle_group_id FROM MuscleGroups WHERE name = 'Chest')),

    ((SELECT exercise_id FROM Exercises WHERE name = 'Deadlift'), (SELECT muscle_group_id FROM MuscleGroups WHERE name = 'Back')),
    ((SELECT exercise_id FROM Exercises WHERE name = 'Deadlift'), (SELECT muscle_group_id FROM MuscleGroups WHERE name = 'Legs')),

    ((SELECT exercise_id FROM Exercises WHERE name = 'Squat'), (SELECT muscle_group_id FROM MuscleGroups WHERE name = 'Legs')),

    ((SELECT exercise_id FROM Exercises WHERE name = 'Shoulder Press'), (SELECT muscle_group_id FROM MuscleGroups WHERE name = 'Shoulders')),

    ((SELECT exercise_id FROM Exercises WHERE name = 'Bicep Curl'), (SELECT muscle_group_id FROM MuscleGroups WHERE name = 'Arms')),

    ((SELECT exercise_id FROM Exercises WHERE name = 'Crunch'), (SELECT muscle_group_id FROM MuscleGroups WHERE name = 'Abs'));
