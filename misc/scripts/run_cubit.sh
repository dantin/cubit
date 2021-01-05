#!/bin/bash
set -e

DEPLOY_DIR=/home/david/Documents/cubit

cd "${DEPLOY_DIR}"

./wait-for 127.0.0.1:3306 -t 1000 -- echo 'MySQL is up'
exec ./cubit \
	-c ./data/prod.yml
