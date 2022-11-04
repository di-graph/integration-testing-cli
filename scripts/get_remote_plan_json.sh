#!/usr/bin/bash
echo "Here"
JSON_FILEPATH="/tmp/plan_json.json"
RUN_URL=`echo $1 | grep -Eo 'https://[^ >]+' | head -1`
RUN_ID=`echo $RUN_URL | awk -F/ '{print $NF}'`
PLAN_ID=`curl -N -s \
--header "Authorization: Bearer $2" \
https://app.terraform.io/api/v2/runs/$RUN_ID | python3 -c "import sys, json; print(json.load(sys.stdin)['data']['relationships']['plan']['data']['id'])"`
curl -L \
--header "Authorization: Bearer $2" \
--header "Content-Type: application/vnd.api+json" \
https://app.terraform.io/api/v2/plans/$PLAN_ID/json-output-redacted > $JSON_FILEPATH
# ls -al $JSON_FILEPATH
echo $JSON_FILEPATH