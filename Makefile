default: help

help: # Prints available commands
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

build: # Build docker image for rinha compiler
	@docker build -t rinha .

hello: # Run rinha program to hello example
	@docker run -it rinha examples/print.rinha.json

showcase: # Run rinha program to showcase example
	@docker run -it rinha examples/showcase.rinha.json

run: # Run rinha program to file /var/rinha/source.rinha.json
	@docker run --cpus=2 --memory=2gb rinha examples/source.rinha.json

ifdef file
    bench: # Run time of rinha program to file argument file=xpto.json
			@time docker run -it rinha $(file)
else
    bench: # Run time of rinha program to showcase example
			@time docker run -it rinha examples/showcase.rinha.json
endif