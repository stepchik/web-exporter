FROM golang:latest
RUN mkdir /app
ADD go.mod /app/
ADD go.sum /app/
ADD main.go /app/
WORKDIR /app
RUN go build -o main .
ENTRYPOINT ["/app/main"," -a", "127.0.0.1:5555", "-f", "sites.json"]
