#!/bin/bash

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

source $dir/pushApi.sh
source $dir/pushFlow.sh
source $dir/pushSecrets.sh
source $dir/pushSidecar.sh
source $dir/pushInitPod.sh
source $dir/pushIsolate.sh
source $dir/pushUI.sh
