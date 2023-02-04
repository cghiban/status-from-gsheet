#!/bin/sh

docker build --no-cache --rm -t gsheet-status-multistage -f Dockerfile.multistage .
