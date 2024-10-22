#!/bin/bash

# Check if $DIREKTIV_UI_BACKEND environment variable is valid host and port string
if [ -z "$DIREKTIV_UI_BACKEND" ]; then
    echo "The environment variable DIREKTIV_UI_BACKEND is not present or empty."
fi
pattern="^(http|https)://([a-zA-Z0-9.-]+)(:[1-9][0-9]{0,4})?(/.*)?$"
if ! [[ $DIREKTIV_UI_BACKEND =~ $pattern ]]; then
      echo "The environment DIREKTIV_UI_BACKEND variable is not a valid http(s) host and port >$DIREKTIV_UI_BACKEND<";
      exit 1;
fi
# Substitute env vars in /etc/nginx/conf.d/default.conf
var=$(echo "$DIREKTIV_UI_BACKEND" | sed 's/\//\\\//g')
sed -i "s/{DIREKTIV_UI_BACKEND}/$var/g" /etc/nginx/conf.d/default.conf



# Check if $DIREKTIV_UI_PORT environment variable is valid port string
if [ -z "$DIREKTIV_UI_PORT" ]; then
    echo "The environment variable DIREKTIV_UI_PORT is not present or empty."
fi
pattern="^([1-9][0-9]{0,4})$"
if ! [[ $DIREKTIV_UI_PORT =~ $pattern ]]; then
      echo "The environment DIREKTIV_UI_PORT variable is not a valid hostname and port >$DIREKTIV_UI_PORT<";
      exit 1;
fi
# Substitute env vars in /etc/nginx/conf.d/default.conf
sed -i "s/{DIREKTIV_UI_PORT}/${DIREKTIV_UI_PORT}/g" /etc/nginx/conf.d/default.conf

cat /etc/nginx/conf.d/default.conf

# Start Nginx
exec nginx -g "daemon off;"
