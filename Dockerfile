FROM golang:1.26 AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o climaCEP ./cmd

FROM scratch
WORKDIR /app
COPY --from=build /app/climaCEP .
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["./climaCEP"]

