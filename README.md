# GYMBRO
GYMBRO API REPO

*... got inspiration from Nikolay Tuzov (just learning...)*

## Dependencies
1) **cleanenv** - parsing configuration
2) **slog** - logging
3) **sqlite3** - DB
4) **chi** - router

## Note
Trying to make pet project for gym rats. Hope this project will teach me a lot :) Trying to understand and comment code (*using my English skills*) as much as I can for future me. *(Starting 18th Jun 2024)*

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
│   ├───config
│   │       config.go == Config handler (getting parameters from file)
│   │
│   └───storage == Common things for all possible storages (not only sqlite)
│   │       storage.go
│   │
│   └───sqlite == Code only related to sqlite storage (init ...)
│           sqlite.go
│
└───storage == Folder where sqlite3 database(.db file) are located
```
