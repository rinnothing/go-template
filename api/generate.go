package api

//go:generate go tool oapi-codegen -generate=types -package=gen -o gen/types.gen.go openapi.yml
//go:generate go tool oapi-codegen -generate=server,strict-server -package=gen -o gen/server.gen.go openapi.yml
//go:generate go tool oapi-codegen -generate=client -package=gen -o gen/client.gen.go openapi.yml
