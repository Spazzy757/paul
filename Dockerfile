# builder image
FROM golang:1.14-buster as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -o paul ./cmd/paul

# generate clean, final image for end users
FROM gcr.io/distroless/base-debian10
COPY --from=builder /build/paul /
CMD ["/paul"]
