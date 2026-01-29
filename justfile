set dotenv-load := true

default:
  just --list

[group('goose')]
[working-directory: 'sql/schema']
goose-up:
    goose postgres $DB_URL up

[working-directory: 'sql/schema']
goose-down:
    goose postgres $DB_URL down

[group('sqlc')]
sql-go:
    sqlc generate

list-dirs:
    ls
