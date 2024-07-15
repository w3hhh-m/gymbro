# GYMBRO
GYMBRO API REPO. Pet-project that uses the following technologies: Chi router, PostgreSQL DB, slog logger, ...

*... got inspiration from Nikolay Tuzov (just learning...)*

## Dependencies
1) **cleanenv** - parsing configuration
2) **slog** - logging
3) **pgx, pgxpool** - DB
4) **chi** - router
5) **chi/render** - manage HTTP request / response
6) **validator** - validator :0

## Note
Trying to make pet project for gym rats. Hope this project will teach me a lot :) Trying to understand and comment code (*using my English skills*) as much as I can for future me. *(Starting 18th Jun 2024)*

## Environment variables
1) **CONFIG_PATH** - path to config.yaml file. for local env in this project - ./config/local.yaml
2) **STORAGE_PATH** - path to postgresql database. example: postgres://username:password@host:port/dbname

## Structure
```
├───cmd == Folder where the main commands are located (such as running app)
│   └───gymbro
│           main.go == Main project file
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
    │   │   ├───record == Handlers for records
    │   │   │   ├───delete
    │   │   │   │       delete.go
    │   │   │   │
    │   │   │   ├───get
    │   │   │   │       get.go
    │   │   │   │
    │   │   │   └───save
    │   │   │           save.go
    │   │   │
    │   │   └───response == Common response things for all handlers
    │   │           response.go
    │   │
    │   └───middleware == Custom middlewares
    │       └───logger == Logger for router (got request, took 1ms, etc.)
    │               logger.go
    │
    ├───prettylogger == Pretty logs for local env
    │       prettylogger.go
    │
    └───storage
        │   storage.go == Common things for all possible storages (not only postgres)
        │
        └───postgresql == Code only related to postgresql storage
                postgresql.go

```
