# builder image
FROM golang:1.14-buster as builder

ARG PRIVATE_KEY
ARG SECRET_KEY
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN mkdir /secrets
RUN CGO_ENABLED=0 GOOS=linux go build -o paul ./cmd/paul
RUN echo $PRIVATE_KEY > /secrets/paul-private-key
RUN echo $SECRET_KEY > /secrets/paul-secret-key

# generate clean, final image for end users
FROM gcr.io/distroless/base-debian10
COPY --from=builder /build/paul /
COPY --from=builder /secret /
CMD ["/paul"]
