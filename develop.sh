#!/bin/bash

docker run -d -e POSTGRES_PASSWORD=postgres -p 5432:5432 --name godep-postgres postgres:9.6-alpine
