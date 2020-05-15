version="0.3.2"
buildid=$(dbus-uuidgen)
dockerTag="localhost:32000/lightpeer-clientjob:$version-$buildid"
docker build -t $dockerTag -f examples/client/Dockerfile .
docker push $dockerTag

curPath=$(pwd)

cd examples/client/k8s
~/go/bin/kustomize edit set image localhost:32000/lightpeer-clientjob=$dockerTag
kubectl apply -k .
cd $curPath 
