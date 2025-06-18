FROM golang:1.24-alpine

RUN apk add --no-cache bash git postgresql-client
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
RUN cd backend && go build -o ../main .
COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh"]