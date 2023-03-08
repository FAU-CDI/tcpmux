FROM alpine as permission

# create www-data
RUN set -x ; \
  addgroup -g 82 -S www-data ; \
  adduser -u 82 -D -S -G www-data www-data && exit 0 ; exit 1

# build the server
FROM golang as build

# build the app
ADD . /app/
WORKDIR /app/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/tcpmux /app/cmd/tcpmux/

# add it into a scratch image
FROM scratch
WORKDIR /

# add the user
COPY --from=permission /etc/passwd /etc/passwd
COPY --from=permission /etc/group /etc/group

# add the app
COPY --from=build /app/tcpmux /tcpmux

# and set the entry command
EXPOSE 8000
USER www-data:www-data
ENTRYPOINT ["/tcpmux"]