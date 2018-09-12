FROM scratch

ADD ./pretty-deps .

ENTRYPOINT ["/pretty-deps"]
