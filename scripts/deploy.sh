#!/usr/bin/env bash

while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    -s|--stack)
    AWS_STACK="$2"
    shift
    shift
    ;;
    -b|--bucket)
    AWS_BUCKET="$2"
    shift
    shift
    ;;
esac
done

echo "AWS STACK  = ${AWS_STACK}"
echo "AWS BUCKET = ${AWS_BUCKET}"

sam deploy --stack-name "${AWS_STACK}" --s3-bucket "${AWS_BUCKET}" --capabilities CAPABILITY_IAM CAPABILITY_AUTO_EXPAND
