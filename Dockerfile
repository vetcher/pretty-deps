FROM golang:1.11 as backend

RUN go get github.com/gin-gonic/gin
RUN go get github.com/gin-contrib/cors
RUN go get github.com/vetcher/pretty-deps

WORKDIR /go/src/github.com/vetcher/pretty-deps
RUN go build -o pretty-deps ./webapp/main.go

FROM node:alpine as frontend

COPY --from=backend /go/src/github.com/vetcher/pretty-deps /go/src/github.com/vetcher/pretty-deps
WORKDIR /go/src/github.com/vetcher/pretty-deps/static
RUN npm install
RUN npm run build

FROM alpine

COPY --from=frontend /go/src/github.com/vetcher/pretty-deps/static/bundle.css ./static/bundle.css
COPY --from=frontend /go/src/github.com/vetcher/pretty-deps/static/bundle.js ./static/bundle.js
COPY --from=backend /go/src/github.com/vetcher/pretty-deps/pretty-deps ./pretty-deps

RUN pwd
RUN ls -all
RUN ls ./static/ -all

ENTRYPOINT ["/pretty-deps"]
EXPOSE 9000
