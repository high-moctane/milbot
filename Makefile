.PHONY: build run clean raspi

build:
	go build -o milbot

run: build
	./milbot

clean:
	rm -f ./milbot
	rm -f ./milbot-raspi

raspi:
	GOOS=linux GOARCH=arm GOARM=6 go build -o milbot-raspi
