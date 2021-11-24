# Clients-emails project.


## Description.
Microservice that stores clients and send them emails based on mailing ID. 

The service uses Go and PostgreSQL.

Both DB and app are containerized and can be run using `docker-compose`.

## Concerns
1. There are no test for every file because I found it pointless. One example test that covers 
   every tested case is enough for presentation purposes.
2. Cursor - I've implemented simple cursor that bases on `where` and `order by` statements.
   It's simple and fast enough for that project. Along with index on `id` column
   even with large amount of data I should be fast enough.
   
   If needed, we could use PostgreSQL builtin cursor functionality: [cursor](https://www.postgresql.org/docs/9.2/plpgsql-cursors.html)
   Also we could use other approach of API pagination like that: [pagination](https://ignaciochiazzo.medium.com/paginating-requests-in-apis-d4883d4c1c4c#:~:text=Most%20of%20the%20use%20cases,%2C%20and%20Cursor%2Dbased%20Pagination.).
3. `Creating customer entries should be idempotent` - for simplicity I've used `UNIQUE INDEX` to ensure that.
   In more complex scenario I'd consider using locks on db because with multiple 
   instance of that app we could encounter problems.
   
## Local setup

Make sure that you have go installed.

1. Install dependencies.
```shell
go mod download & go mod vendor
```

2. Crate DB if not exists.
```shell
docker run --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=postgres POSTGRES_DB=vodeno POSTGRES_USER=postgres -d postgres:14
```

3. Run an app.
```shell
go run cmd/main.go
```

4. Or use docker-compose setup. It will build docker image of an app and pull postgres image. App will run on `8080` port.
```shell
docker-compose up
```