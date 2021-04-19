#!/usr/bin/env bash

sam build -t modules/store/template.yaml -b .aws-sam/build/store && \
sam build -t modules/api/template.yaml -b .aws-sam/build/api && \
sam build -t modules/core/template.yaml -b .aws-sam/build/core && \
sam build -t template.yaml -b .aws-sam/build/root
