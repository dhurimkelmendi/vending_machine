#This is only meant to be used in development. DO NOT USE THIS TO SERVE PRODUCTION SERVER
serve: main.go
	go run main.go

db_upgrade: main.go
	go run main.go migrate up

db_reset: main.go
	go run main.go migrate reset

db_downgrade: main.go
	go run main.go migrate down

test:
	go test -count=1 -parallel 1 -v ./...
