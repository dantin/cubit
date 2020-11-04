#!/bin/sh
set -e

# first arg is `-so` or `--some-option`
# default is server
if [ "${1#-}" != "$1" ]; then
    set -- server "$@"
fi

if [ "$(id -u)" = '0' ]; then
    find . \! -user admin -exec chown admin:admin '{}' +
    if [ "$1" = 'server' ]; then
        exec gosu admin /app/cubit -c /app/prod.yml "${@:2}"
    fi
fi

exec "$@"
