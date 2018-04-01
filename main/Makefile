run:
	@go run main.go

build:
	@go tool compile -S main.go >> main.s 
	@rm main.o
clean:
	@rm main.s