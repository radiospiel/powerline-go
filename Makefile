default: install

build:
	go build -o powerline-go .

install: build
	cp powerline-go /usr/local/bin

