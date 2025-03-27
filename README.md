# Pricetra GraphQL API

## Docker
We run our development Postgres database in a Docker container, so make sure docker is installed (see [Docker desktop](https://docs.docker.com/desktop/)).

### Start docker container
```
$ make docker-container
```
This will also map the default `5432` port to a new `5431` port to avoid collisions. Check to see if the new postgres db is running on port `5431`.

## Start Server
After completing all the steps from above, we can finally start our server.

```
$ make run
```

Assuming all the generated files remain the same, this will spin up a server on port `8080`, and will also perform all the migration as defined in `./database/migrations`.
There should be at least 2 routes for development. `/playground` and `/graphql`. You can use the GraphQL Playground feature by navigating to http://localhost:8080/playground.

In order to run the server in watch mode run `make watch` instead.


## Jet
We use [go-jet/jet](https://github.com/go-jet/jet) to handle all database related queries, insertions, updates, and deletes. Jet uses an active DB connection to generate the appropriate models, and functions needed for the query builder. To run this, use the command `make jet`. Rerun this command after your migrations have been set (see Migrations section)

## GraphQL
We use [99designs/gqlgen](https://github.com/99designs/gqlgen) to handle all GraphQL related tasks. This library allows us to define all gql resolvers, directives, types, etc. and generates Go files to use or implement. To generate queries, mutations, types, etc. use the command `make gql`. This command must be run after any changes made to the `.graphql` files inside `./graph/`.

## Migrations
For creating an entity migration run
```
$ make create-migration fileName=name_of_your_migration
```
This will create a new file under `./database/migrations` with a UNIX timestamp. Add your migration to the SQL file and run the server to set the migration.

Also note that after applying the migration a new schema will be created named `atlas_schema_revisions` which will hold the `atlas_schema_revisions` table listing all the successful migrations.
