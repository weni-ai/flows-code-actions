```
╔═╗┬  ┌─┐┬ ┬┌─┐  ╔═╗┌─┐┌┬┐┌─┐  ╔═╗┌─┐┌┬┐┬┌─┐┌┐┌┌─┐
╠╣ │  │ ││││└─┐  ║  │ │ ││├┤   ╠═╣│   │ ││ ││││└─┐
╚  ┴─┘└─┘└┴┘└─┘  ╚═╝└─┘─┴┘└─┘  ╩ ╩└─┘ ┴ ┴└─┘┘└┘└─┘
```

[Postman Collection](./doc/Code%20Actions.postman_collection_v3.json)

[WORK IN PROGRESS]

### requirements:

* go 1.21.4

* mongodb running local


### how to run

download dependencies:
```
go mod download -x
```

install air:
```
go install github.com/cosmtrek/air@latest
```

run:

```
air -d
```
