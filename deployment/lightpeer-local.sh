CGO_ENABLED=0 GOOS=linux go build -o ./bin/lightpeer ./src/lightpeer/

version="0.1.0"
buildid=$(dbus-uuidgen)
dockerTag="localhost:32000/lightpeer:$version-$buildid"
docker build -t $dockerTag -f Dockerfile .
docker push $dockerTag

curPath=$(pwd)

cd k8s
~/go/bin/kustomize edit set image localhost:32000/lightpeer=$dockerTag
kubectl apply -k .
cd $curPath 
