#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
set -x

body="{
    \"request\": {
        \"message\": \"Triggerd by ${TRAVIS_REPO_SLUG}#${TRAVIS_JOB_NUMBER}\",
        \"branch\": \"master\"
    }
}"

request=$(
    curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Accept: application/json" \
        -H "Travis-API-Version: 3" \
        -H "Authorization: token ${TRAVIS_API_TOKEN}" \
        -d "$body" \
        https://api.travis-ci.com/repo/projectriff%2Fbuilder/requests
)
request_id=`echo $request | jq '.request.id'`
sleep 5
request=$(
    curl -s \
        -H "Accept: application/json" \
        -H "Travis-API-Version: 3" \
        -H "Authorization: token ${TRAVIS_API_TOKEN}" \
        https://api.travis-ci.com/repo/projectriff%2Fbuilder/request/${request_id}
)

echo "Triggered a new builder build"
echo -e "View results at https://travis-ci.com/projectriff/builder/builds/`echo $request | jq -r '.builds[0].id'`"