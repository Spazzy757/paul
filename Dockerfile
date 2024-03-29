# builder image
FROM golang:1.18-buster as builder

RUN mkdir /build
ADD . /build/
WORKDIR /build

RUN CGO_ENABLED=0 GOOS=linux go build -o paul ./cmd/paul

ARG PRIVATE_KEY
ARG SECRET_KEY
ENV GITHUB_PRIVATE_KEY=$PRIVATE_KEY
ENV GITHUB_SECRET_KEY=$SECRET_KEY
RUN mkdir /secrets
RUN echo ${GITHUB_PRIVATE_KEY} | base64 -d > /secrets/paul-private-key
RUN echo ${GITHUB_SECRET_KEY} > /secrets/paul-secret-key

FROM node:lts-alpine as web-builder
RUN mkdir /web
WORKDIR /app
COPY web/package*.json .
RUN yarn install
COPY web/ .
RUN yarn build
RUN mv dist /web/dist

# generate clean, final image for end users
FROM gcr.io/distroless/base-debian10
COPY --from=builder /build/paul /
COPY --from=builder /secrets /secrets
COPY --from=web-builder /web /web
CMD ["/paul"]
