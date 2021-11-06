init:
	go mod vendor
	
build:
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags netgo -o dist/ec2-boot-cinema-linux ec2-boot-cinema.go
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -tags netgo -o dist/ec2-boot-cinema-macos ec2-boot-cinema.go

test:
	env CGO_ENABLED=0
