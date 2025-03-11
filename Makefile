APP_NAME=api-status-checker


run:
	go run main.go

build: 
	go build -o $(APP_NAME) main.go

clean:
	rm -f $(APP_NAME)