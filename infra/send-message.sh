#!/bin/bash

while true; do
  listening=$(docker logs valheim-server | grep "Server is now listening")
  
  if [ -z "$listening" ]; then
    logger -t valheim-server-check "Server is not listening yet"  # Logs to syslog
  else
    ACCESS_TOKEN=$(curl -H "Metadata: true" "http://169.254.169.254/metadata/identity/oauth2/token?resource=https://storage.azure.com/&api-version=2018-02-01" | jq -r .access_token)
    
    STORAGE_ACCOUNT_NAME="discordevents"
    QUEUE_NAME="messages"
    MESSAGE="listening"
    BASE64_MESSAGE=$(echo -n "$MESSAGE" | base64)
    
    curl -X POST "https://${STORAGE_ACCOUNT_NAME}.queue.core.windows.net/${QUEUE_NAME}/messages?timeout=60" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "x-ms-version: 2018-03-28" \
    -H "Content-Type: application/xml" \
    --data "<QueueMessage><MessageText>${BASE64_MESSAGE}</MessageText></QueueMessage>"
  fi

  sleep 5
done
