# rmq_docker_demo

### Install 
    - Docker
    - Golang

### Setup
go to ./consumer and ./publisher and run
```
    go mod tidy
```

### Run
```
    docker-compose up --build
```

### Test
POST
```
    curl -X POST -H "Content-Type: application/json" -d '{"id":2,"title":"New Album","artist":"New Artist"}' http://localhost:8080/album
```

GET
```
    http://localhost:8080/album?id=1
```

### Stop
```
    docker-compose down
```