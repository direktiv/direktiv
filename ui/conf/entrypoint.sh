#!/bin/bash

# Check if $UI_BACKEND environment variable is valid host and port string
if [ -z "$UI_BACKEND" ]; then
    echo "The environment variable UI_BACKEND is not present or empty."
fi
pattern="^(http|https)://([a-zA-Z0-9.-]+):([1-9][0-9]{0,4})(/.*)?$"
if ! [[ $UI_BACKEND =~ $pattern ]]; then
      echo "The environment UI_BACKEND variable is not a valid http(s) host and port >$UI_BACKEND<";
      exit 1;
fi
# Substitute env vars in /etc/nginx/conf.d/default.conf
var=$(echo "$UI_BACKEND" | sed 's/\//\\\//g')
sed -i "s/{UI_BACKEND}/$var/g" /etc/nginx/conf.d/default.conf



# Check if $UI_PORT environment variable is valid port string
if [ -z "$UI_PORT" ]; then
    echo "The environment variable UI_PORT is not present or empty."
fi
pattern="^([1-9][0-9]{0,4})$"
if ! [[ $UI_PORT =~ $pattern ]]; then
      echo "The environment UI_PORT variable is not a valid hostname and port >$UI_PORT<";
      exit 1;
fi
# Substitute env vars in /etc/nginx/conf.d/default.conf
sed -i "s/{UI_PORT}/${UI_PORT}/g" /etc/nginx/conf.d/default.conf


cat /etc/nginx/conf.d/default.conf

# Start Nginx
exec nginx -g "daemon off;"
