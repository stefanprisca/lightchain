DockerHost = localhost:32000

build-lightpeer:
	docker build -t $(DockerHost)/lightpeer:example ../../
	docker push $(DockerHost)/lightpeer:example

deploy-example:
	kubectl create ns lightpeer

	docker build -t $(DockerHost)/persister:0.1.0 ./persister
	docker push $(DockerHost)/persister:0.1.0
	kubectl apply -f persister/

	docker build -t $(DockerHost)/reader:0.1.0 ./reader
	docker push $(DockerHost)/reader:0.1.0
	kubectl apply -f reader/
