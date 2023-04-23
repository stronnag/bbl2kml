ser2tcp: main.go enumerate_linux.go  enumerate_other.go  go.sum
	go build -ldflags "-s -w"

go.sum: go.mod
	go mod tidy

clean:
	rm -f ser2tcp ser2tcp.exe go.sum
