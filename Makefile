compile: 
	go build -o bin/go-grab.exe .

run: compile
	./bin/go-grab.exe
