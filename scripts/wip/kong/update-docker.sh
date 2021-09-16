#!/bin/sh

rm -rf scripts/wip/kong/proto
mkdir scripts/wip/kong/proto/
mkdir scripts/wip/kong/proto/google/
mkdir -p scripts/wip/kong/proto/tmp
# mkdir -p scripts/wip/kong/proto/pkg/flow/grpc/

# copy proto files
echo "Copying proto files to tmp dir"
# find pkg/flow/grpc -name '*.proto' -exec cp -prv '{}' 'scripts/wip/kong/proto/' ';'
# find pkg/flow/grpc -name '*.proto' -exec cp -prv '{}' 'scripts/wip/kong/proto/pkg/flow/grpc/' ';'
find pkg/flow/grpc -name '*.proto' -exec cat '{}' \; > scripts/wip/kong/proto/tmp/mega.proto
sed '/^import/ d' < scripts/wip/kong/proto/tmp/mega.proto > scripts/wip/kong/proto/tmp/mega2.proto
sed '/^syntax/ d' < scripts/wip/kong/proto/tmp/mega2.proto > scripts/wip/kong/proto/tmp/mega3.proto
sed '/^option/ d' < scripts/wip/kong/proto/tmp/mega3.proto > scripts/wip/kong/proto/tmp/mega4.proto
sed '/^package/ d' < scripts/wip/kong/proto/tmp/mega4.proto > scripts/wip/kong/proto/tmp/mega5.proto

printf '%s\n%s\n' "$(cat scripts/wip/kong/proto.hack)" "$(cat scripts/wip/kong/proto/tmp/mega5.proto)" >scripts/wip/kong/proto/flow.proto
# cat scripts/wip/kong/proto/tmp/mega5.proto >scripts/wip/kong/proto/flow.proto

rm -rf scripts/wip/kong/proto/tmp


cp -r /usr/include/google/api/ scripts/wip/kong/proto/google/
cp -r /usr/include/google/protobuf/ scripts/wip/kong/proto/google/

cd scripts/wip/kong; docker build -t localhost:5000/cong .

docker push localhost:5000/cong

# delete old pods

echo "delete old pods"

kubectl delete pod -l app.kubernetes.io/name=kong-external
# kubectl delete pod -n kong -l app=ingress-kong