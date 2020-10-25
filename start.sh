#!/bin/sh
docker run -d --rm --name test-postgres -e POSTGRES_DB=blocketdb -e POSTGRES_HOST_AUTH_METHOD=trust -p 13000:5432 -v `pwd`/apps/db/init/:/docker-entrypoint-initdb.d/ postgres:9.4-alpine