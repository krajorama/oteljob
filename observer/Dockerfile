FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 go build -o /observer .

FROM gcr.io/distroless/static-debian12
COPY --from=build /observer /observer
EXPOSE 9464
ENTRYPOINT ["/observer"]
