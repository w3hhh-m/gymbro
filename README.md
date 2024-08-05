# GYMBRO

GYMBRO API REPO. Pet-project that uses the following technologies:

1) Chi router
2) PostgreSQL DB
3) slog logger
4) JWT Auth and OAuth
5) DB Migrations
6) Unit and Fuzz test using Testify
7) Mocks for tests with Mockery

*... got inspiration from Nikolay Tuzov (just learning...)*

## Dependencies

1) **cleanenv** - parsing configuration
2) **slog** - logging
3) **pgx, pgxpool** - DB
4) **chi** - router
5) **chi/render** - manage HTTP request / response
6) **validator** - validator :0
7) **bcrypt** - hashing passwords
8) **jwt** - JWT :0
9) **golang-migrate/migrate** - migrations :0
10) **goth, gorilla/sessions** - OAuth
11) **mockery** - mocks :0
12) **testify** - test :0
13) **google/uuid** - uuid generation

## Note

Trying to make pet project for gym rats. Hope this project will teach me a lot :). Kinda first project using golang, so
please dont judge strictly :)

## Environment variables

1) **CONFIG_PATH** - path to config.yaml file. for local env in this project - ./config/local.yaml
2) **STORAGE_PATH** - path to postgresql database. example: postgres://username:password@host:port/dbname
3) **SECRET_KEY** - secret for jwt tokens
4) **GOOGLE_KEY** - client id from console.google.cloud.com for OAuth
5) **GOOGLE_SECRET** - secret from console.google.cloud.com for OAuth

## Migrations

Don't forget to set **STORAGE_PATH** and run `go run ./cmd/migrate --direction=[up|down]`

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
    │   │   │       users_handler_factory.go
    │   │   │       workouts_handler_factory.go
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
    │   │       ├───records
    │   │       │   └───add
    │   │       │           add.go
    │   │       │           add_test.go
    │   │       │
    │   │       ├───scheduler
    │   │       │       scheduler.go
    │   │       │
    │   │       ├───sessions
    │   │       │       sessions.go
    │   │       │
    │   │       └───start
    │   │               start.go
    │   │               start_test.go
    │   │
    │   └───middleware == Custom middlewares
    │       ├───jwt == For JWT Auth
    │       │       jwt.go
    │       │
    │       ├───logger == Logger for router (got request, took 1ms, etc.)
    │       │       logger.go
    │       │
    │       └───workout == For workout session check
    │               workout.go
    │               workout_test.go
    │
    ├───lib
    │   ├───jwt
    │   │       jwt.go
    │   │
    │   └───prettylogger == Pretty logs for local env
    │           prettylogger.go
    │
    └───storage
        │   storage.go == Common things for all possible storages (not only postgres)
        │
        ├───mocks == Mocks for Unit testing handlers
        │       UserRepository.go
        │       WorkoutRepository.go
        │
        └───postgresql == Code only related to postgresql storage
                postgresql.go

```
