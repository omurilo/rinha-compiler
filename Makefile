default: help

 .PHONY: help
help: # Prints available commands
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

.PHONY: build
build: # Build docker image for rinha compiler
	@docker build -t rinha .

.PHONY: hello
hello: # Run rinha program to hello example
	@docker run -it rinha examples/print.json

.PHONY: showcase
showcase: # Run rinha program to showcase example
	@docker run -it rinha examples/showcase.json

.PHONY: run
run: # Run rinha program to file /var/rinha/source.rinha.json
	@docker run -v ./source.rinha.json:/var/rinha/source.rinha.json --cpus=2 --memory=2gb rinha

.PHONY: bench
bench: # Run time of rinha program to showcase example
	@time docker run -it rinha examples/showcase.json