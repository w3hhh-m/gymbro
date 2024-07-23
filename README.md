# GYMBRO
GYMBRO API REPO. Pet-project that uses the following technologies: Chi router, PostgreSQL DB, slog logger, JWT Auth, OAuth, Migrations, ...

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

## Note
Trying to make pet project for gym rats. Hope this project will teach me a lot :) Trying to understand and comment code (*using my English skills*) as much as I can for future me. *(Starting 18th Jun 2024)*

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
    │   │   ├───records == Handlers for records
    │   │   │   ├───delete
    │   │   │   │       delete.go
    │   │   │   │
    │   │   │   ├───get
    │   │   │   │       get.go
    │   │   │   │
    │   │   │   └───save
    │   │   │           save.go
    │   │   │
    │   │   ├───response == Common response things for all handlers
    │   │   │       response.go
    │   │   │
    │   │   └───users == Handlers for users
    │   │       ├───login
    │   │       │       login.go
    │   │       │
    │   │       ├───logout
    │   │       │       logout.go
    │   │       │
    │   │       ├───oauth
    │   │       │       oauth.go
    │   │       │
    │   │       └───register
    │   │               register.go
    │   │
    │   └───middleware == Custom middlewares
    │       └───logger == Logger for router (got request, took 1ms, etc.)
    │               logger.go
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
        └───postgresql == Code only related to postgresql storage
                postgresql.go

```
