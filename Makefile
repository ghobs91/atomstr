PREFIX  ?= /usr/local

atomstr: clean
	GOOS=linux GOARCH=amd64 go build -o atomstr -ldflags="-s -w"

linux-arm:
	GOOS=linux GOARCH=arm go build -o atomstr -ldflags="-s -w"

darwin:	
	GOOS=darwin GOARCH=amd64 go build -o atomstr -ldflags="-s -w"

windows:
	GOOS=windows GOARCH=amd64 go build -o atomstr.exe -ldflags="-s -w"

clean:
	rm -f atomstr

install: 
	install -d $(PREFIX)/bin/
	install -m 755 atomstr $(PREFIX)/bin/atomstr

uninstall:
	rm -f $(PREFIX)/bin/atomstr
