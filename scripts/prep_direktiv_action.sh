#!/bin/bash

direkcli namespaces create test

cat > helloworld.yml <<- EOF
 id: httpget
 functions:
 - id: httprequest
   image: localhost:5000/demo-action
 states:
 - id: getter
   type: action
   action:
     function: httprequest
     input: '{
       "method": "GET",
       "url": "https://jsonplaceholder.typicode.com/todos/1",
     }'
EOF

direkcli workflows create test helloworld.yml

direkcli workflows execute test httpget
