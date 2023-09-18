FROM golang:1.20.7 as builder

WORKDIR /var/app

RUN apt update && apt upgrade -y

RUN git clone https://github.com/omurilo/rinha-compiler.git /var/app

RUN GOOS=linux go build -o rinha /var/app/main.go

RUN cp rinha /usr/local/bin

ENTRYPOINT [ "/bin/bash" ]