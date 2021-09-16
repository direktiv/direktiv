#!/usr/bin/env bash
# This script is does four things in order
# 1) Create workflow that calls knative service (this is so we can trigger a reconstruction)
# 2) Save a snapshot of the current details of a service and its revisions. Push this snapshot to a namespace var
# 3) Delete using kubectl and run the workflow to trigger a reconsturction
# 4) Save a snapshot of the reconstructed details of a service and its revisions. Push this snapshot to a namespace var
# The test.direktion workflow can now compare the two snapshots to see if the service was recreated successfully.

function="namespace-test-request"

# We have to do this in script for now until direktion support knative functions
# TODO: When direktion supports knative functions we can remove this

# We need to sleep here also to wait for service rollout to be done
sleep 20
wf="id: init-script-workflow
functions:
  - id: get
    type: knative-namespace
    service: request
states:
  - id: getter
    type: action
    action:
      function: get
      input:
        method: GET
        url: https://jsonplaceholder.typicode.com/todos/1"

echo "CREATE init-script-workflow workflow"
resp=`curl -s -S -X POST $DIREKTIV_API/api/namespaces/$NAMESPACE/workflows -H "Content-Type: text/yaml" --data-raw "$wf"`
status=$?
if [ $status -ne 0 ]; then exit $status; fi

echo "Saving a revision details snapshot of the $function service to $NAMESPACE namespace variable '$function-original'"
revisions=`curl -s -S $DIREKTIV_API/api/namespaces/$NAMESPACE/functions/$function`
status=$?
if [ $status -ne 0 ]; then exit $status; fi

originalSnapshot=`echo $revisions | jq -r '. | {"name": .name, "namespace": .namespace, "config": .config, "revisions": [.revisions[] | {"name": .name, "image": .image, "minScale": .minScale, "generation": .generation, "traffic": .traffic}]}'`
resp=`curl -s -S -X POST $DIREKTIV_API/api/namespaces/$NAMESPACE/variables/$function-original --data-raw "$originalSnapshot"`
status=$?
if [ $status -ne 0 ]; then exit $status; fi


echo "DELETING KNATIVE SERVICE $function"
kubectl delete services.serving.knative.dev -n direktiv-services-direktiv $function

sleep 5

echo "RECONSTRUCTING KNATIVE SERVICE $function"
resp=`curl -I -s -S -X GET $DIREKTIV_API/api/namespaces/$NAMESPACE/workflows/init-script-workflow/execute?wait=true`
code=`echo "$resp" | head -1 | cut -f2 -d" "`

if [ $code -ne 200 ] ; then
    echo "  init-script-workflow workflow returned unsuccessful status code: $code"
    exit 1
fi

# sleep 5 seconds to give time for service traffic to settle
sleep 20

echo "Saving a revision details snapshot of the $function service to $NAMESPACE namespace variable '$function-reconstructed'"
revisions=`curl -s -S $DIREKTIV_API/api/namespaces/$NAMESPACE/functions/$function`
status=$?
if [ $status -ne 0 ]; then exit $status; fi

reconstructedSnapshot=`echo $revisions | jq -r '. | {"name": .name, "namespace": .namespace, "config": .config, "revisions": [.revisions[] | {"name": .name, "image": .image, "minScale": .minScale, "generation": .generation, "traffic": .traffic}]}'`
resp=`curl -s -S -X POST $DIREKTIV_API/api/namespaces/$NAMESPACE/variables/$function-reconstructed --data-raw "$reconstructedSnapshot"`
status=$?
if [ $status -ne 0 ]; then exit $status; fi