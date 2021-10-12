#!/usr/bin/env bash

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
GREY='\033[0;37m'
NC='\033[0m' # No Color
NL=$'\n'

DIREKTION=$HOME/go/src/github.com/vorteil/direktion/build/direktion
DIREKTIV_API="http://localhost:80"
NAMESPACE=test

suite_dir=${1%/}

echo "Searching for tests in: $suite_dir"

tests=`find $suite_dir -maxdepth 1 -type d ! -name "_*" | tail --lines=+2 | sort`
count=`echo "$tests" | wc -l`
failed=0

init_direktion () {
	if [ -f "$DIREKTION" ]; then
		echo -e "${GREEN}USING DIREKTION FROM GO PATH${NC}"
		return 0
	fi

	gopath=$DIREKTION

	DIREKTION=`which direktion`
	if [ -f "$DIREKTION" ]; then
		echo -e "${GREEN}USING DIREKTION FROM PATH${NC}"
		return 0
	fi

	echo -e "${RED}DIREKTION BINARY NOT FOUND IN SYSTEM PATH OR GO PATH: $gopath ${NC}"
	exit 1
}

build_isolate () {
	isolate=$1
	tag="localhost:5000/${isolate##*/}:test"

	echo "isolate =========== $isolate"

	cmd=$(docker build -t $tag $isolate 2>&1)
	status=$?
	if ((status)); then
		echo -e "    ${RED}error ${NC} encounted building isolate; Dumping output";

		# add padding
		cmd="$(echo "$cmd" | sed 's/^/    /')"
		echo >&2 "$cmd"
		return $status
	fi

	cmd=$(docker push $tag 2>&1)
	status=$?
	if ((status)); then
		echo -e "    ${RED}error ${NC} encounted pushing isolate; Dumping output";

		# add padding
		cmd="$(echo "$cmd" | sed 's/^/    /')"
		echo >&2 "$cmd"
		return $status
	fi

}


build_isolates () {
	dir=$1
	isolates=`find $dir -maxdepth 2 -type d -wholename '*_Isolates/isolate-*'`
	is=`echo "$isolates" | wc -l`
	if [ ! -z "$isolates" ]
	then
		echo -e "${BLUE}Building Isolates${NC}"
	fi

	idx=0
	for isolate in $isolates; do 
		let idx++
		echo "  building isolate ($idx/$is): $isolate..."
		build_isolate $isolate 
		status=$?
		if [ $status -ne 0 ]; then return $status; fi
	done
}

create_services () {
	dir=$1
	services=`find $dir -maxdepth 2 -type f -wholename '*_Services/*.json'`
	is=`echo "$services" | wc -l`
	if [ ! -z "$services" ]
	then
		echo -e "${BLUE}Creating Services${NC}"
	fi

	idx=0
	for service in $services; do 
		let idx++
		if [[ $service == *.global.json ]]; then
			echo "  building services ($idx/$is): $service..."
			build_service $service "global"
			status=$?
			if [ $status -ne 0 ]; then return $status; fi
		else
			echo "  building services ($idx/$is): $service..."
			build_service $service "ns"
			status=$?
			if [ $status -ne 0 ]; then return $status; fi
		fi
	done
}

build_service () {
	service=$1
	scope=$2

	name=`basename $service`
	name=${name%.*}
	name=${name%.*}
	revisions=`cat $service | jq '.revisions'`
	traffic=`cat $service | jq '.traffic'`
	rev=1
	skiped="false"

	if [[ $scope == "ns" ]]; then
		echo "service =========== namespace-$NAMESPACE-$name"
		for row in $(echo "${revisions}" | jq -r '.[] | @base64'); do
			echo "revision =========== namespace-$NAMESPACE-$name-0000$rev"
			if [ "$rev" -eq "1" ]; then
				# Create Service
				data=`echo ${row} | base64 --decode | jq --arg namespace "$NAMESPACE" --arg name "$name" '. + {"namespace": $namespace, "name": $name}'`
				resp=`curl -s -S -X POST $DIREKTIV_API/api/functions/namespaces/$NAMESPACE --data-raw "$data"`
				status=$?
				if [ $status -ne 0 ]; then return $status; fi
				code=`echo $resp | jq '.Code'`
				if [[ $code == "2" ]]; then
					echo -e "service =========== namespace-$NAMESPACE-$name already exists - ${BLUE}skipping all revisions${NC}"
					skiped="true"
					break
				fi
			else
				# Create revision
				data=`echo ${row} | base64 --decode | jq '. + {"trafficPercent": 100}'`
				resp=`curl -s -S -X POST $DIREKTIV_API/api/functions/namespaces/$NAMESPACE/function/$name --data-raw "$data"`
				status=$?
			
				if [ $status -ne 0 ]; then return $status; fi
			fi

			# increment rev
			((rev=rev+1))
		done

		# sleep for 1 sec if more than service revisions creation was not skipped
		if [[ $skiped == "false" ]]; then
			sleep 1
		fi

		revisionPrefix="namespace-$NAMESPACE-$name-"
		trafficData=`echo $traffic | jq --arg revisionPrefix "$revisionPrefix" '{values: [.[] | {"percent": .percent, "revision": ($revisionPrefix + .genSuffix)}]}'`
		echo "traffic =========== $(echo $trafficData | jq -c ".values")"
		resp=`curl -s -S -X PATCH $DIREKTIV_API/api/functions/namespaces/$NAMESPACE/function/$name --data-raw "$trafficData"`
		status=$?
		
		if [ $status -ne 0 ]; then return $status; fi
	else
		echo -e "service =========== global-$name ${RED} NOT IMPLEMENTED ${NC}";
	fi
}

push_variable () {
	var=$1
	workflow=$2

	name=`basename $var`
	name=${name%.*}

	if [ -z "$workflow" ]
	then
		resp=`curl -s -S -X PUT $DIREKTIV_API/api/namespaces/$NAMESPACE/variables/$name --data-binary "@$var"`
		status=$?
		
		if [ $status -ne 0 ]; then return $status; fi
	else
		resp=`curl -s -S -X PUT $DIREKTIV_API/api/namespaces/$NAMESPACE/tree/$workflow?op=set-var&var=$name --data-binary "@$var"`
		status=$?
		
		if [ $status -ne 0 ]; then return $status; fi
	fi

	

}

push_namespace_variables () {
	dir=$1
	vars=`find $dir -maxdepth 2 -type f -wholename '*_NamespaceVariables/*.variable'`
	iv=`echo "$vars" | wc -l`
	if [ ! -z "$vars" ]
	then
		echo -e "${BLUE}Pushing Namespace Variables${NC}"	
	fi

	idx=0
	for var in $vars; do 
		let idx++
		echo "  pushing namespace variable ($idx/$iv): $var..."
		push_variable $var $workflow
		status=$?
		if [ $status -ne 0 ]; then return $status; fi
	done
}

push_workflow_variables () {
	test=$1
	workflow=$2
	var_pattern="${workflow}Variable"

	vars=`find $test -maxdepth 1 -type f -name "*.$var_pattern"`
	iv=`echo "$vars" | wc -l`

	idx=0
	for var in $vars; do 
		let idx++
		echo "  pushing workflow variable ($idx/$iv): $var..."
		push_variable "$var" "$workflow"
		status=$?
		if [ $status -ne 0 ]; then return $status; fi
	done
}

compile_script () {
	script=$1

	wf=`$DIREKTION compile $script 2>/dev/null`
	status=$?
	if [ $status -ne 0 ]; then 
		echo -e "    ${RED}error${NC} encounted compiling script; run the following command to reproduce:";
		echo "      $DIREKTION compile $script";
		return $status
	fi

	# upload
	tmp="/tmp/test-workflow.yaml"
	echo "$wf" > $tmp

	filename="$(basename $script)"
	filename=${filename%.direktion}
	
	resp=`curl -s -S -X PUT $DIREKTIV_API/api/namespaces/$NAMESPACE/tree/$filename?op=create-workflow -H "Content-Type: text/yaml" --data-binary "@$tmp"`
	status=$?
	
	if [ $status -ne 0 ]; then return $status; fi

	id=`echo "$resp" | jq -r ".id"`
	resp=`curl -s -S -X PUT $DIREKTIV_API/api/namespaces/$NAMESPACE/tree/$filename?op=set-workflow-event-logging -H "Content-Type: text/yaml" --data "{\"logger\": \"$filename\"}"`
	status=$?

	if [ $status -ne 0 ]; then return $status; fi

	rm $tmp

	push_workflow_variables $(dirname $script) $id
}

compile_scripts () {
	test=$1
	scripts=`find $test -maxdepth 1 -type f -name '*.direktion'`
	sc=`echo "$scripts" | wc -l`
	
	idx=0
	for script in $scripts; do 
		let idx++
		echo "  compiling script ($idx/$sc): $script..."
		compile_script $script 
		echo "script =========== $script"
		status=$?
		if [ $status -ne 0 ]; then return $status; fi
	done

}

run_test_init_script () {
	test=$1
	script=$test/init.sh
	if [ -f "$script" ]; then
		echo "init-start =========== running script (This may take awhile) $script"
		init_output=`DIREKTIV_API=$DIREKTIV_API NAMESPACE=$NAMESPACE bash -c $script`
		status=$?
		echo "$init_output" | while IFS= read -r line ; do echo -e "    ${GREY}$line${NC}"; done
		if [ $status -ne 0 ]; then return $status; fi
		echo "init-end =========== running script $script"
		return 0
	fi
}

wipe_namespace () {

	# wipe nodes 
	resp=`curl -s -S -X GET $DIREKTIV_API/api/namespaces/$NAMESPACE/tree/`
	paths=`echo "$resp" | jq -r '.children.edges[].node.path' 2>/dev/null`
	status=$?
	if [ $status -eq 0 ]; then 
		for path in $paths; do
			resp=`curl -s -S -X DELETE $DIREKTIV_API/api/namespaces/$NAMESPACE/tree$path`
		done
	fi

	return 0
}

perform_test () {
	test=$1

	if [ ! -d "$test" ]; then echo -e "  ${RED}error${NC} no directory for test: $test"; return 1; fi

	wipe_namespace

	compile_scripts $test
	status=$?
	if [ $status -ne 0 ]; then return $status; fi

	run_test_init_script $test
	status=$?
	if [ $status -ne 0 ]; then return $status; fi

	resp=`curl -I -s -S -X GET $DIREKTIV_API/api/namespaces/$NAMESPACE/tree/test?op=wait`
	code=`echo "$resp" | head -1 | cut -f2 -d" "`

	if [ $code -ne 200 ] ; then
		echo "  test workflow returned unsuccessful status code: $code"
		return 1
	fi

	if echo "$resp" | grep -q "direktiv_errorcode"; then
		echo "  test workflow returned with state errors:"
		echo -e "  ${RED}$(echo "$resp" | grep "direktiv_errorcode")${NC}"
		echo -e "  ${RED}$(echo "$resp" | grep "direktiv_errormsg")${NC}"
		return 1
	fi

	return 0
}

report_results () {
	status=$1
	if [ $status -eq 0 ]; then 
		echo -e "  ${GREEN}SUCCESS${NC}"
	else 
		echo -e "  ${RED}FAIL${NC}"
		let failed++
	fi
}

perform_tests () {
	i=0
	for test in $tests; do 
		let i++
		echo -e "${BLUE}Running test${NC} ($i/$count): $test..."
		perform_test $test 
		status=$?
		report_results $status
	done
}

individuals=0

perform_individual_tests () {
	shift
	for test in "$@"
	do
		let individuals++
		echo -e "${BLUE}Running test${NC} $individuals: $test..."
		perform_test "$suite_dir/$test"
		status=$?
		report_results $status
	done
}

init_namespace () {

	resp=`curl -I -s -S -X PUT $DIREKTIV_API/api/namespaces/$NAMESPACE`
	code=`echo "$resp" | head -1 | cut -f2 -d" "`

	if [ $code -ne 409 ] && [ $code -ne 200 ]; then
		echo "$resp"
		return 1
	else 
		echo "Created namespace: $NAMESPACE"
		return 0
	fi

	wipe_namespace
	status=$?

	return $status

}

init_secrets () {

	resp=`curl -s -S -X PUT $DIREKTIV_API/api/namespaces/test/secrets/URL_SECRET --data 'https://docs.direktiv.io'`
	code=`echo "$resp" | head -1 | cut -f2 -d" "`
	grpcCode=`echo $resp | jq '.Code'`
	if [[ $grpcCode == "2" ]]; then
		grpcMsg=`echo $resp | jq '.Message'`
		echo -e "${BLUE}Skipping secret creation${NC}: $grpcMsg"
		return 0
	fi

	if  [ $code -ne 409 ] && [ $code -ne 200 ]; then 
		echo "$resp"
		return 1
	else
		echo "Created secret: URL_SECRET"
		return 0
	fi 

    status=$?
	return $status
}

if [ "$1" = "" ]; then 
	echo "usage: $0 TESTSUITE_DIR [TESTNAME]..."
	exit 1
fi

init_direktion

init_namespace
create_services $1

init_secrets
status=$?
if [ $status -ne 0 ]; then exit $status; fi
build_isolates $1
push_namespace_variables $1

perform_individual_tests $@

if [ $individuals -eq 0 ]; then 
	perform_tests
fi

if [ $failed -eq 0 ]; then 
	echo "All tests completed successfully."
	exit 0
else
	echo "$failed tests failed."
	exit 1
fi

