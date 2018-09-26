FROM golang:1.11 as backend

RUN go get github.com/gin-gonic/gin github.com/gin-contrib/cors github.com/vetcher/pretty-deps

WORKDIR /go/src/github.com/vetcher/pretty-deps
RUN CGO_ENABLED=0 GOOS=linux go build -o pretty-deps ./webapp/main.go

FROM node:alpine as frontend

COPY --from=backend /go/src/github.com/vetcher/pretty-deps /go/src/github.com/vetcher/pretty-deps
WORKDIR /go/src/github.com/vetcher/pretty-deps/static
RUN npm install
RUN npm run build

FROM scratch

COPY --from=frontend /go/src/github.com/vetcher/pretty-deps/static/bundle.css ./static/bundle.css
COPY --from=frontend /go/src/github.com/vetcher/pretty-deps/static/bundle.js ./static/bundle.js
COPY --from=frontend /go/src/github.com/vetcher/pretty-deps/static/index.html ./static/index.html
COPY --from=backend /go/src/github.com/vetcher/pretty-deps/pretty-deps ./pretty-deps
LABEL description="pretty-deps"

ENTRYPOINT ["/pretty-deps"]
EXPOSE 9000
