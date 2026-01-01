#!/usr/bin/env bash

docker compose -f /data/docker-compose.yaml down

docker compose -f /data/docker-compose.yaml pull

docker compose -f /data/docker-compose.yaml up -d
