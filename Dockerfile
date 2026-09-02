# syntax=docker/dockerfile:1

# --- build stage ---
FROM golang:1.27-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -o /uhkm \
    .

# --- runtime stage ---
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /uhkm /uhkm

ENTRYPOINT ["/uhkm"]
