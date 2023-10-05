FROM  golang:1.20-alpine as builder
ENV CGO_ENABLED=0
WORKDIR /
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
RUN go build -o /rinha .

FROM gcr.io/distroless/base-debian11
LABEL inspiredBy="Erick Amorim <ericklima.ca@yahoo.com>"
LABEL maintainer="Murilo Alves <hi@omurilo.dev>"
COPY --from=builder /rinha /rinha
COPY --from=builder /examples /examples
ENTRYPOINT [ "/rinha" ]