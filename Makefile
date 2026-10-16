.PHONY: staticcheck

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck@v0.8.2 ./...
