build:
	go build -o bin/strawk strawk.go
install:
	go install strawk.go
clean:
	rm -rf bin
	rm -rf program.awk
	rm -rf test
test: build
	./tests/test.zsh
