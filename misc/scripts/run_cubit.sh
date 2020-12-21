#!/bin/bash
set -e

DEPLOY_DIR=/home/david/Documents/cubit

cd "${DEPLOY_DIR}"

exec ./cubit \
	-c ./data/prod.yml
