# GYMBRO

GYMBRO API REPO. Project that uses the following technologies:

1) Chi router
2) PostgreSQL DB
3) slog logger
4) JWT Auth and OAuth
5) DB Migrations
6) Unit tests using Testify
7) Mocks for tests with Mockery
8) Redis

## Dependencies

1) **cleanenv** - parsing configuration
2) **slog** - logging
3) **pgx, pgxpool** - PostgreSQL
4) **chi** - router
5) **chi/render** - manage HTTP request / response
6) **validator**
7) **bcrypt**
8) **jwt**
9) **golang-migrate/migrate**
10) **goth, gorilla/sessions** - OAuth
11) **mockery**
12) **testify**
13) **google/uuid**
14) **redis/go-redis/v9** - Redis

## Note

Trying to make pet project for gym rats. Hope this project will teach me a lot :). Kinda first project using golang, so
please dont judge strictly :)

## Environment variables

1) **CONFIG_PATH** - path to config.yaml file. for local env in this project - ./config/local.yaml
2) **STORAGE_PATH** - path to postgresql database. example: postgres://username:password@host:port/dbname
3) **SECRET_KEY** - secret for jwt tokens
4) **GOOGLE_KEY** - client id from console.google.cloud.com for OAuth
5) **GOOGLE_SECRET** - secret from console.google.cloud.com for OAuth
6) **REDIS_PATH** - path to redis
7) **REDIS_PASSWORD** - password for redis :0

## Migrations

Don't forget to set **STORAGE_PATH** and run `go run ./cmd/migrate --direction=[up|down]`

## Key Features

1. **Configuration & Initialization**:  
   At startup, the application loads all necessary configurations from environment variables and config files, establishes connections to PostgreSQL and Redis, and initializes the logger.

2. **Router & Middleware**:  
   The Chi router is set up with handler factories to inject dependencies, and essential middlewares like RequestID, URLFormat, and Recoverer are integrated. Custom middleware logs request details, including execution time.

3. **OAuth & JWT Authentication**:  
   Google OAuth is configured for user authentication. Protected routes require a valid JWT token, ensuring secure access to user-specific features.

4. **Session Management**:  
   Active workout sessions are managed via Redis. When adding or modifying workout records, the application checks for an active session to ensure records are associated with the correct workout.

5. **Unit Testing & Transactions**:  
   The application is tested using mocks. Transactions are used for complex database operations to ensure data consistency.

6. **Session Scheduler**:  
   A session scheduler periodically checks Redis for inactive sessions, automatically ends them, and saves workout data to the database if necessary.

This streamlined setup ensures that app runs efficiently, securely manages user sessions, and reliably handles workout data.

## Structure

```
├───cmd == Folder where the main commands are located (such as running app)
│   ├───gymbro
│   │       main.go == Main project file
│   │
│   └───migrate == Database migrations
│       │   main.go
│       │
│       └───migrations
│               1_init.down.sql
│               1_init.up.sql
│               2_fill.down.sql
│               2_fill.up.sql
│
├───config == Folder where config files are located
│       local.yaml
│
└───internal == Folder with internal handlers 
    ├───config
    │       config.go == Config handler (getting parameters from file)
    │
    ├───http-server
    │   ├───handlers == Server handlers :0
    │   │   ├───factory == Abstract factory creation pattern
    │   │   │       abstract_handler_factory.go
    │   │   │       middlewares_handler_factory.go
    │   │   │       records_handler_factory.go
    │   │   │       users_handler_factory.go
    │   │   │       workouts_handler_factory.go
    │   │   │
    │   │   ├───records == Handlers for records
    │   │   │   ├───add
    │   │   │   │       add.go
    │   │   │   │       add_test.go
    │   │   │   │
    │   │   │   └───delete
    │   │   │           delete.go
    │   │   │           delete_test.go
    │   │   │
    │   │   ├───response == Common response things for all handlers
    │   │   │       response.go
    │   │   │
    │   │   ├───users == Handlers for users
    │   │   │   ├───login
    │   │   │   │       login.go
    │   │   │   │       login_test.go
    │   │   │   │
    │   │   │   ├───logout
    │   │   │   │       logout.go
    │   │   │   │       logout_test.go
    │   │   │   │
    │   │   │   ├───oauth
    │   │   │   │       oauth.go
    │   │   │   │
    │   │   │   └───register
    │   │   │           register.go
    │   │   │           register_test.go
    │   │   │
    │   │   └───workouts == Handlers for workouts
    │   │       ├───end
    │   │       │       end.go
    │   │       │       end_test.go
    │   │       │
    │   │       ├───get
    │   │       │       get.go
    │   │       │       get_test.go
    │   │       │
    │   │       └───start
    │   │               start.go
    │   │               start_test.go
    │   │
    │   └───middleware == Custom middlewares
    │       ├───jwt == For JWT Auth
    │       │       jwt.go
    │       │       jwt_test.go
    │       │
    │       ├───logger == Logger for router (got request, took 1ms, etc.)
    │       │       logger.go
    │       │
    │       └───workout == For workout session check
    │               workout.go
    │               workout_test.go
    │
    ├───lib
    │   ├───jwt == Custom JWT getter, generator, validator
    │   │       jwt.go
    │   │
    │   ├───points == Points calc
    │   │       points.go
    │   │
    │   ├───prettylogger == Pretty logs for local env
    │   │       prettylogger.go
    │   │
    │   └───validation == Custom validation messages
    │           validation.go
    │
    └───storage
        │   storage.go == Common things for all possible storages (not only postgres)
        │
        ├───mocks == Mocks for Unit testing handlers
        │       SessionRepository.go
        │       UserRepository.go
        │       WorkoutRepository.go
        │
        ├───postgresql == Code only related to PostgreSQL storage
        │        postgresql.go
        │
        └───redis == Code only related to Redis storage
                redis.go
```
