FROM golang:1.24
WORKDIR /app
COPY . /app
RUN go build -o main main.go
EXPOSE 8080
CMD ["./main"]
