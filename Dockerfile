FROM alpine:3.12

MAINTAINER david chengjie.ding@gmail.com

# add user and group first to make sure their IDs get assigned consistently, regardless of whatever dependencies get added.
RUN addgroup -S admin && adduser -S admin -G admin

# grap gosu for easy step-down from root
# https://github.com/tianon/gosu/releases
ENV GOSU_VERSION 1.11
RUN set -eux; \
	\
	apk add --no-cache --virtual .gosu-deps \
		ca-certificates \
		dpkg \
		gnupg \
	; \
	\
	dpkgArch="$(dpkg --print-architecture | awk -F- '{ print $NF }')"; \
	wget -O /usr/local/bin/gosu "https://github.com/tianon/gosu/releases/download/$GOSU_VERSION/gosu-$dpkgArch"; \
	wget -O /usr/local/bin/gosu.asc "https://github.com/tianon/gosu/releases/download/$GOSU_VERSION/gosu-$dpkgArch.asc"; \
	\
# verify the signature
	export GNUPGHOME="$(mktemp -d)"; \
	gpg --batch --keyserver hkps://keys.openpgp.org --recv-keys B42F6819007F00F88E364FD4036A9C25BF357DD4; \
	gpg --batch --verify /usr/local/bin/gosu.asc /usr/local/bin/gosu; \
	command -v gpgconf && gpgconf --kill all || :; \
	rm -rf "$GNUPGHOME" /usr/local/bin/gosu.asc; \
	\
# clean up fetch dependencies
	apk del --no-network .gosu-deps; \
	\
	chmod +x /usr/local/bin/gosu; \
# verify that the binary works
	gosu --version; \
	gosu nobody true

RUN apk update && \
        apk add --no-cache bash

# app related
RUN mkdir /app
WORKDIR /app

COPY entrypoint.sh /usr/local/bin
COPY wait-for /app/wait-for
COPY cubit /app/cubit
COPY data/cert/server.crt /app/cert/server.crt
COPY data/cert/server.key /app/cert/server.key
COPY data/prod.yml /app/prod.yml

# Make scripts runnable
RUN chmod +x /usr/local/bin/entrypoint.sh
RUN chmod +x /app/wait-for
RUN chmod +x /app/cubit

ENTRYPOINT ["entrypoint.sh"]

CMD ["server"]

EXPOSE 5222
