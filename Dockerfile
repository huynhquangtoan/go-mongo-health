ARG GO_VERSION=1.20.6

FROM golang:${GO_VERSION}-alpine  AS builder

# git is required to fetch go dependencies
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build \
  -installsuffix 'static' \
  -o /home/mongo-health/app .

# Final stage: the running container.
FROM alpine AS final

# Copy CA cert
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Import the user and group files from the first stage.
#COPY --from=builder /user/group /user/passwd /etc/

# Import the compiled executable from the first stage.
COPY --from=builder /home/mongo-health/app /home/mongo-health/app

# Run the compiled binary.
ENTRYPOINT ["/home/mongo-health/app"]