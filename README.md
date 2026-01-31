# ⛰️ pimpmypack

> PimpMyPack is a set of backend APIs dedicated to CRUD operations on hiking equipment inventories and packing lists.  
> It should be used in conjunction with any frontend candidates.  
> It could replace [Lighterpack](https://lighterpack.com/) if this project dies (because it's not maintained anymore)  

## PimpMyPack API

The server is based on [Gin Framework](https://github.com/gin-gonic/gin) and provides endpoints to manage Accounts, Inventories & Packs

A dedicated API documentation is available [here](https://pmp-dev.alki.earth/swagger/index.html).

## Setup for local development

### 1. clone this repo

```shell
git clone git@github.com:Angak0k/pimpmypack.git
```

### 3. Start a local postgres database

The app need a local DB.

You need to use docker to start a postgres database:

```shell
docker run --name pmp_db \
    -d -p 5432:5432 \
    -e POSTGRES_PASSWORD=pmp1234 \
    -e POSTGRES_USER=pmp_user \
    -e POSTGRES_DB=pmp_db postgres:17
```

**Note:** PostgreSQL 17 is required for this project.

### 4. Configure the environment

Pimpmypack app read its conf from the environment and/or `.env` file.

The simplest way is to:

* copy the `.env.sample` file to `.env`
* customize the values in the `.env` file to match your setup

### 5. Start the API server

```shell
go build . && ./pimpmypack
```

## Run tests

```shell
go test ./...
```

or with verbose mode

```shell
go test -v ./...
```
