build: ; docker build --tag pretty-deps:poc .

run: ; docker run -p 127.0.0.1:9000:9000 --name pretty-deps pretty-deps:poc -addr=0.0.0.0:9000
