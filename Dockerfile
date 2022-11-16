# This is a multi-stage Dockerfile and requires >= Docker 17.05
# https://docs.docker.com/engine/userguide/eng-image/multistage-build/
#FROM gobuffalo/buffalo:v0.16.9 as builder
FROM golang as builder

ENV GO111MODULE on
ENV GOPROXY http://proxy.golang.org

RUN mkdir -p /src/github.com/tcarreira/roaw2020
WORKDIR /src/github.com/tcarreira/roaw2020

# Installing Node 12
RUN curl -sL https://deb.nodesource.com/setup_12.x | bash
RUN apt-get update && apt-get install -y nodejs python
RUN npm install --global yarn

# this will cache the npm install step, unless package.json changes
ADD package.json .
ADD yarn.lock .
RUN yarn install --no-progress

RUN go install github.com/gobuffalo/buffalo/buffalo@v0.16.9

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download
RUN go get github.com/gobuffalo/flect@v0.2.1

ADD . .
RUN buffalo build --static -o /bin/app

FROM alpine
RUN apk add --no-cache curl
RUN apk add --no-cache bash
RUN apk add --no-cache ca-certificates

WORKDIR /bin/

COPY --from=builder /bin/app .

# Uncomment to run the binary in "production" mode:
# ENV GO_ENV=production

# Bind the app to 0.0.0.0 so it can be seen from outside the container
ENV ADDR=0.0.0.0

EXPOSE 3000

# Uncomment to run the migrations before running the binary:
# CMD /bin/app migrate; /bin/app
CMD exec /bin/app

