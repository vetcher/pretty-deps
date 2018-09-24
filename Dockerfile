FROM golang:1.11 as backend

WORKDIR /go/src/github.com/vetcher/pretty-deps
RUN go build -o pretty-deps ./webapp/main.go
ADD ./pretty-deps .

FROM node:alpine as frontend

WORKDIR /go/src/github.com/vetcher/pretty-deps/static
RUN npm install
RUN npm run build

FROM scratch

COPY --from=frontend /go/src/github.com/vetcher/pretty-deps/static/bundle.css ./static/bundle.css
COPY --from=frontend /go/src/github.com/vetcher/pretty-deps/static/bundle.js ./static/bundle.js
COPY --from=backend /go/src/github.com/vetcher/pretty-deps/pretty-deps ./pretty-deps

ENTRYPOINT ["/pretty-deps"]
