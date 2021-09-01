.PHONY: backend build clean test

backend:	
	go build

build:
	docker build -t feitaemcasa .

clean:
	$(RM) backend.zip