# gomigrate
A database agnostic migration tool for go projects.

Go projects tend to all be slightly different and have their own idiosyncrasies. 

## Goals

* Timestamp-based database migrations.
* migration generators based on project-specific templates
* support up and down migrations for green path and optional rollback support
* single command to run all pending migrations that have not already run successfully. state stored in database with hook db-specific hook methods or "drivers".
* post-up and post-down migrations for tasks to do after the deploy is complete.
* allow re-running migrations if explicitly requested
* "migrator install" command to install templates
- verify function to verfiy the integrity of the migration and report on unexpected differences
- migrations are easy to write tests for
- write docker-based tests




## Instalation

1. `go get github.com/ssoroka/gomigrate`
2. `go install github.com/ssoroka/gomigrate/migrate`
3. add ~/go/bin to your path if it's not already. Something along the lines of: 
```bash
  $ echo export PATH=\$PATH:`go env GOPATH`/bin >> ~/.bash_profile
  $ export PATH=$PATH:`go env GOPATH`/bin
```
3. To install templates in your current project: `migrate install`

## Usage

- Create a new migration: `migrate new`
- Run all pending migration scripts: `migrate up`

## Running Project Tests

To be added

## Writing Migration Tests

Examples coming