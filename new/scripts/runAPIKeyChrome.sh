#!/bin/bash

# this script is used to test the api key in development 
# it avoids CORS warnings if the UI and server run on different servers
# requires chrome obviously

google-chrome --disable-web-security --user-data-dir="[/tmp]"
