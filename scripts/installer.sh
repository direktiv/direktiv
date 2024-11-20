#!/usr/bin/env -S bash --posix

# TODO: curl -fsSL get.docker.com -o get-docker.sh && sh get-docker.sh

# TODO: ensure skips still notice version mismatches
DEV=${DEV:-false}
VERBOSE=${VERBOSE:-false}
WITH_MONITORING=${WITH_MONITORING:-false}

DIREKTIV_CONFIG=${DIREKTIV_CONFIG:-direktiv.yaml}

INSTALL_CONTOUR_VERSION="${INSTALL_CONTOUR_VERSION:-v1.11.1}"
INSTALL_K3S_VERSION="${INSTALL_K3S_VERSION:-v1.28.1+k3s1}"
INSTALL_KNATIVE_VERSION="${INSTALL_KNATIVE_VERSION:-v1.11.7}"

# NOTE: DIREKTIV_VERSION is not exposed because the script is not necessarily
#   expected to work across multiple versions. The script should be re-published
#   for each release with this value updated.
DIREKTIV_VERSION="v0.8.0"
 
check_help() {
    # Print usage information if any '-h' or '--help' flag was given. This 
    # script doesn't do flags, but a user won't know that without good help
    # information.
    for a in $@; do
        if [[ ( $a == "--help") ||  $a == "-h" ]]
        then 
            print_usage
            exit 0
        fi 
    done

    # Print usage if the number of arguments wasn't exactly one.
    if [[ $# -ne 1 ]]
    then
        print_usage
        exit 0
    fi 
}

print_usage() {
  cat <<EOF
Usage: $0 [command]

    This script is used to install Direktiv ${DIREKTIV_VERSION}. 
    
    All options have sensible defaults, so if you don't know what you're doing 
    you probably just need to use the 'install_all' command.

Options:
    INSTALL_CONTOUR_VERSION     version     Set the Contour version.
    INSTALL_K3S_VERSION         version     Set the k3s version.
    INSTALL_KNATIVE_VERSION     version     Set the knative version.
    
    DEV                         boolean     Install from local dev environment.
    VERBOSE                     boolean     Enable verbose script output.
    WITH_MONITORING             boolean     Enable grafana-stack installation

Commands:
    all         Install everything.
    uninstall   Uninstall everything.

EOF
}

log() {
    echo $@
}

verbose() {
    log $@
}

assert_success() {
    status=$1
    msg=$2
    output=$3

    if [ $status -ne 0 ]
    then 
        echo "$msg:"
        echo $output
        exit $status
    fi
}

check_is_k3s_installed() {
    if ! command -v kubectl &> /dev/null 
    then 
        return 1
    fi
}

check_is_k3s_version_match() {
    if ! check_is_k3s_installed
    then
        return 1
    fi

    output=`kubectl version 2>&1 3>/dev/null`
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    output=`echo "$output" | head -n 1`

    if [ "$output" == "Client Version: ${INSTALL_K3S_VERSION}" ]
    then 
        return 0
    fi

    return 1
}   

install_k3s() {    
    if check_is_k3s_version_match
    then
        verbose "Skipping k3s install step: ${INSTALL_K3S_VERSION} already installed."

        return 0 
    fi

    log "Installing k3s..."

    # TODO: proxy setup option

    output=`curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION=$INSTALL_K3S_VERSION sh -s - --disable traefik --write-kubeconfig-mode=644 2>&1 | tee /dev/fd/3`
    assert_success $? "Failed to install k3s" "$output"

    # TODO: customizable k3s configuration

    log "Successfully installed k3s ${INSTALL_K3S_VERSION}."
}

uninstall_k3s() {    
    if ! check_is_k3s_installed
    then
        return 0
    fi

    output=`k3s-uninstall.sh 2>&1 | tee /dev/fd/3`
    assert_success $? "Failed to uninstall k3s" "$output"

    log "Successfully uninstalled k3s."
}

check_is_database_already_installed() {
    kubectl -n postgres get service direktiv-cluster-pods &>/dev/null
    
    return $?
}   

add_db_operator_helm_repo() {
    output=`helm repo add percona https://percona.github.io/percona-helm-charts/ 2>&1 | tee /dev/fd/3`
    assert_success $? "Failed to add percona helm repository" "$output"
}

add_fluentbit_helm_repo() {
    output=`helm repo add fluent-bit https://fluent.github.io/helm-charts 2>&1 | tee /dev/fd/3`
    assert_success $? "Failed to add fluentbit helm repository" "$output"
}

check_is_db_namespace_exists() {
    kubectl get namespace postgres &>/dev/null

    return $?
}   

create_kubernetes_db_namespace() {
    if check_is_db_namespace_exists
    then
        return 0 
    fi

    output=`kubectl create namespace postgres 2>&1 | tee /dev/fd/3`
    assert_success $? "Failed to create database namespace" "$output"
}

helm_install_db_operator() {    
    output=`helm -n postgres status pg-operator 2>&1 | tee /dev/fd/3`
    status=$?

    if [ $status -eq 0 ] 
    then
        # NOTE: this is where we would put an upgrade if we supported that
        #       output=`helm upgrade -n postgres pg-operator percona/pg-operator --wait 2>&1 | tee /dev/fd/3`

        return 0
    fi

    output=`helm install -n postgres pg-operator percona/pg-operator --wait 2>&1 | tee /dev/fd/3`
    assert_success $? "Failed to install postgres operator" "$output"
}

apply_database_configuration() {
    output=`kubectl apply -f ./scripts/kubernetes/install/db/basic.yaml 2>&1 | tee /dev/fd/3`
    assert_success $? "Failed to apply database configuration" "$output"

    # TODO: customizable database configuration
}

await_database() {
    # this loop is necessary because kubernetes often hasn't created the deployment before this first call
    while ! output=`kubectl -n postgres get deployment direktiv-cluster-pgbouncer &>/dev/null`
    do
        sleep 1
    done

    kubectl -n postgres rollout status deployment/direktiv-cluster-pgbouncer --timeout=1s &>/dev/null
    status=$?

    if [ $status -eq 0 ]
    then 
        return 0
    fi

    log "Waiting for database to be live. This could take a while..."

    output=`kubectl -n postgres rollout status deployment/direktiv-cluster-pgbouncer 2>&1 | tee /dev/fd/3`
    assert_success $? "Database deployment may have failed" "$output"

    output=`kubectl -n postgres wait --timeout=300s --for=condition=ready pod -l "postgres-operator.crunchydata.com/instance-set=instance1"`
    assert_success $? "Database deployment may have failed" "$output"
}

update_database_password() {
    # TODO: fix this complaining about a missing role
    return 0

    output=`kubectl get secrets -n postgres direktiv-cluster-pguser-direktiv -o 'go-template={{index .data "password"}}' | base64 --decode`
    assert_success $? "Failed to get database password" "$output"

    password="$output"

    output=`kubectl get pods -n postgres -l "postgres-operator.crunchydata.com/instance-set=instance1" -o=json | jq -r .items[0].metadata.name`
    assert_success $? "Failed to get pod name" "$output"

    pod="$output"

    commands="ALTER USER direktiv WITH PASSWORD '${password}';

\q
"

    output=`echo "$commands" | kubectl exec -n postgres --stdin "${pod}" database -- psql`
    assert_success $? "Failed to exec psql into pod" "$output"
}

install_db() {
    if check_is_database_already_installed
    then
        verbose "Skipping database install step: already installed."

        return 0 
    fi

    log "Installing database..."

    add_db_operator_helm_repo
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    add_fluentbit_helm_repo
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    create_kubernetes_db_namespace
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    helm_install_db_operator
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    apply_database_configuration
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    await_database
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    update_database_password
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    log "Successfully installed database."

    # TODO: version lock the database in a more transparent way
}

check_is_knative_already_installed() {
    kubectl get namespace knative-serving &>/dev/null

    return $?
}

install_knative_operator() {
    output=`kubectl apply -f https://github.com/knative/operator/releases/download/knative-${INSTALL_KNATIVE_VERSION}/operator.yaml 2>&1 | tee /dev/fd/3`
    assert_success $? "Failed to apply knative operator" "$output"
}

create_knative_namespace() {
    output=`kubectl create namespace knative-serving 2>&1 | tee /dev/fd/3`
    assert_success $? "Failed to create knative-serving namespace" "$output"
}

apply_knative_configuration() {
    conf="./scripts/kubernetes/install/knative/basic.yaml"
    if [ "$DEV" != "true" ]; then 
        conf="./scripts/kubernetes/install/knative/basic.yaml/${conf}"
    fi

    log "Applying knative configuration from: ${conf}"

    output=`kubectl apply -f ${conf} 2>&1 | tee /dev/fd/3`
    assert_success $? "Failed to apply knative configuration" "$output"
}

install_contour() {
    output=`kubectl apply --filename https://github.com/knative/net-contour/releases/download/knative-${INSTALL_CONTOUR_VERSION}/contour.yaml 2>&1 | tee /dev/fd/3`
    assert_success $? "Failed to install Contour" "$output"
}

prune_contour_namespaces() {
    output=`kubectl delete namespace contour-external 2>&1 | tee /dev/fd/3`
    assert_success $? "Failed to prune unnecessary Contour namespace" "$output"
}

install_knative() {
    if check_is_knative_already_installed
    then
        verbose "Skipping knative install step: already installed."

        return 0 
    fi

    log "Installing knative..."

    install_knative_operator
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    create_knative_namespace
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    apply_knative_configuration
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    install_contour 
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    prune_contour_namespaces
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    log "Successfully installed knative."
}

check_is_direktiv_already_installed() {
    kubectl get namespace direktiv &>/dev/null

    return $?
}

create_direktiv_namespace() {
    if check_is_direktiv_already_installed
    then
        return 0
    fi

    output=`kubectl create namespace direktiv 2>&1 | tee /dev/fd/3`
    assert_success $? "Failed to create direktiv namespace" "$output"
}

add_direktiv_helm_repo() {
    output=`helm repo add direktiv https://charts.direktiv.io 2>&1 | tee /dev/fd/3`
    assert_success $? "Failed to add Direktiv helm repository" "$output"
}


generate_direktiv_config() {
    echo "Generating Direktiv configuration: ${DIREKTIV_CONFIG}"

    backend_image="direktiv/direktiv"
    frontend_image="direktiv/frontend"

    if test -f "${DIREKTIV_CONFIG}"
    then
        echo "Found existing config file. Will only overwrite database fields."
    else 
        touch $DIREKTIV_CONFIG

        if [ "$DEV" == "true" ]; then 
        DIREKTIV_VERSION="latest"
        backend_image="direktiv"
        frontend_image="frontend"

        cat <<EOF > $DIREKTIV_CONFIG
registry: localhost:5000
image: ${backend_image}
tag: ${DIREKTIV_VERSION}

EOF
        fi

        cat <<EOF >> $DIREKTIV_CONFIG

frontend:
  image: "${frontend_image}"
  tag: "${DIREKTIV_VERSION}"

EOF

        if [ "$WITH_MONITORING" == "true" ]; then
            cat <<EOF >> $DIREKTIV_CONFIG
flow:
  debug: true
fluent-bit:
  install: true
  envFrom:
    - secretRef:
        name: direktiv-fluentbit
  config:
    inputs: |
      [INPUT]
          Name                    tail
          Path                    /var/log/containers/*flow*.log,/var/log/containers/*direktiv-sidecar*.log
          Mem_Buf_Limit           5MB
          Skip_Long_Lines         Off
          Tag                     input
          multiline.parser        cri, docker
          Refresh_Interval        1
          Buffer_Max_Size         64k
    outputs: |
      [OUTPUT]
          name                    pgsql
          match                   flow.*
          port                    ${PG_PORT}
          table                   fluentbit
          user                    ${PG_USER}
          database                ${PG_DB_NAME}
          host                    ${PG_HOST}
          password                ${PG_PASSWORD}

      [OUTPUT]
          Name                    loki
          Match                   *
          Host                    loki.default
          Port                    3100
          Labels                  job=fluentbit
          Line_Format             json
    filters: |
      [FILTER]
          Name                    rewrite_tag
          Match                   input
          Rule                    $log ^.*"track":"([^"]*).*$ flow.$1 true
      [FILTER]
          Name parser
          Match *
          Parser json
          Key_Name log
          Reserve_Data on
opentelemetry:
  # -- opentelemetry address where Direktiv is sending data to
  address: "tempo.default:4317"
  # -- installs opentelemtry agent as sidecar in flow
  enabled: true
  # -- config for sidecar agent
  agentconfig: |
    receivers:
      otlp:
        protocols:
          grpc:
          http:
    exporters:
      otlp:
        endpoint: "tempo.default:4317"
        insecure: true
        sending_queue:
          num_consumers: 4
          queue_size: 100
        retry_on_failure:
          enabled: true
      logging:
        loglevel: debug
    processors:
      batch:
      memory_limiter:
        # Same as --mem-ballast-size-mib CLI argument
        ballast_size_mib: 165
        # 80% of maximum memory up to 2G
        limit_mib: 400
        # 25% of limit up to 2G
        spike_limit_mib: 100
        check_interval: 5s
    extensions:
      zpages: {}
    service:
      extensions: [zpages]
      pipelines:
        traces:
          receivers: [otlp]
          processors: [memory_limiter, batch]
          exporters: [logging, otlp]
EOF
        else
            cat <<EOF >> $DIREKTIV_CONFIG
flow:
  logging: console
EOF
        fi
    fi

    sed -i '/database:/,+6 d' $DIREKTIV_CONFIG
    echo "database:
  host: \"$(kubectl get secrets -n postgres direktiv-cluster-pguser-direktiv -o 'go-template={{index .data "host"}}' | base64 --decode)\"
  password: \"$(kubectl get secrets -n postgres direktiv-cluster-pguser-direktiv -o 'go-template={{index .data "password"}}' | base64 --decode)\"" >> $DIREKTIV_CONFIG
}

install_monitoring() {
    echo "Installing monitoring components..."
    kubectl apply -f scripts/install-monitoring.yaml
    helm repo add grafana https://grafana.github.io/helm-charts
    helm repo add fluent https://fluent.github.io/helm-charts
    helm repo update
    helm upgrade --install tempo grafana/tempo

    echo "Monitoring components installed successfully."
}

install_direktiv() {
    # NOTE: no point every skipping this step

    # TODO: actually, we should skip this if in DEV mode. Instead, deleting pods so they can be automatically re-pulled. Probably.

    generate_direktiv_config
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    log "Installing Direktiv..."

    create_direktiv_namespace
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    # add_direktiv_helm_repo
    # if [ $? -ne 0 ]
    # then 
    #     exit 1
    # fi

    chart="direktiv/direktiv"

    if [ "$DEV" == "true" ]; then 
        chart="charts/direktiv"

        # TODO: when did this become important, and why?
        output=`helm repo add nginx https://kubernetes.github.io/ingress-nginx 2>&1 | tee /dev/fd/3`
        if [ $? -ne 0 ] 
        then 
            assert_success 1 "Failed to add nginx helm repo" "$output"
        fi

        output=`helm dependency update $chart 2>&1 | tee /dev/fd/3`
        if [ $? -ne 0 ] 
        then 
            assert_success 1 "Failed to build helm dependencies" "$output"
        fi

        # TODO: when did this become important, and why?
        output=`helm dependency build $chart 2>&1 | tee /dev/fd/3`
        if [ $? -ne 0 ] 
        then 
            assert_success 1 "Failed to build helm dependencies" "$output"
        fi
    fi

    # TODO: refactor the logic that follows to avoid [[]]

    output=`helm -n direktiv status direktiv 2>&1 | tee /dev/fd/3`
    if [ $? -eq 0 ] 
    then
        output=`helm upgrade -n direktiv -f ${DIREKTIV_CONFIG} direktiv ${chart} --wait 2>&1 | tee /dev/fd/3`
        if [ $? -ne 0 ] 
        then 
            assert_success 1 "Failed to upgrade Direktiv" "$output"
        fi
        
    elif [[ "$output" == *"not found"* ]]
    then
        output=`helm install -n direktiv -f ${DIREKTIV_CONFIG} direktiv ${chart} --wait 2>&1 | tee /dev/fd/3`
        if [ $? -ne 0 ] 
        then 
            assert_success 1 "Failed to install Direktiv" "$output"
        fi

    else 
        assert_success 1 "Failed to check Direktiv status" "$output"
    fi

    log "Successfully installed Direktiv."

    # TODO: await
    # TODO: UI options
    # TODO: dev options
}

install_all() {
    if [ "$DEV" == "true" ]; then 
        log "Installing with developer settings..."
    fi

    install_k3s
    if [ $? -ne 0 ]
    then 
        exit 1
    fi
    
    # TODO: linkerd option

    install_db
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    install_knative
    if [ $? -ne 0 ]
    then 
        exit 1
    fi

    if [ "$WITH_MONITORING" == "true" ]; then
        install_monitoring
    fi

    install_direktiv 
    if [ $? -ne 0 ]
    then 
        exit 1
    fi
}

uninstall_all() {
    uninstall_k3s

    log "Successfully uninstalled Direktiv."
}

#
# SCRIPT BEGINS HERE
#

command=$1

check_help "$@"

# fd 3 and pipefail are used to allow us to suppress output
set -o pipefail
if [ "$VERBOSE" == "true" ]; then 
    exec 3>&1
else 
    exec 3>/dev/null
fi

case "$command" in
    "all")
        install_all
        exit $?
    ;;
    "help") 
        print_usage
        exit 0
    ;;
    "uninstall")
        uninstall_all
        exit $?
    ;;
esac

# If we get here no commands were matched.
print_usage
