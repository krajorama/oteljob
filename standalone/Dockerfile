FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 go build -o /oteljob .

FROM gcr.io/distroless/static-debian12
COPY --from=build /oteljob /oteljob
EXPOSE 9464
ENTRYPOINT ["/oteljob"]
