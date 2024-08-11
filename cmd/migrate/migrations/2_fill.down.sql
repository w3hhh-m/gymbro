DELETE FROM ExerciseMuscleGroups
WHERE exercise_id IN (
    SELECT exercise_id FROM Exercises WHERE name IN ('Bench Press', 'Deadlift', 'Squat', 'Shoulder Press', 'Bicep Curl', 'Crunch')
)
  AND muscle_group_id IN (
    SELECT muscle_group_id FROM MuscleGroups WHERE name IN ('Chest', 'Back', 'Legs', 'Shoulders', 'Arms', 'Abs')
);

DELETE FROM Exercises
WHERE name IN ('Bench Press', 'Deadlift', 'Squat', 'Shoulder Press', 'Bicep Curl', 'Crunch');

DELETE FROM MuscleGroups
WHERE name IN ('Chest', 'Back', 'Legs', 'Shoulders', 'Arms', 'Abs');
