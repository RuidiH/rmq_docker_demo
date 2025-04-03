# Rabbitmq Docker Demo

### Install 
[Docker](https://www.docker.com/products/docker-desktop/)
[Golang](https://go.dev/dl/)

Note: For malware detection issue when installing Docker, see [here](https://github.com/docker/for-mac/issues/7527)

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

### Resource

[Claude](https://claude.northeastern.edu/login/)
