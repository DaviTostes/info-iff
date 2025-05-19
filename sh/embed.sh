#!/bin/sh

DIR="/home/toast/iff_bot/chunks"
URL="https://bot.mediumblue.space/embedding"

for file in "$DIR"/*; do
  if [ -f "$file" ]; then
    echo "Uploading $file..."
    curl -X POST "$URL" \
      -F "file=@$file" \
      -H "Content-Type: multipart/form-data"
  fi
done
