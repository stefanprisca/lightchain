#// Copyright 2020 Stefan Prisca
# //
# // Licensed under the Apache License, Version 2.0 (the "License");
# // you may not use this file except in compliance with the License.
# // You may obtain a copy of the License at
# //
# //     http://www.apache.org/licenses/LICENSE-2.0
# //
# // Unless required by applicable law or agreed to in writing, software
# // distributed under the License is distributed on an "AS IS" BASIS,
# // WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# // See the License for the specific language governing permissions and
# // limitations under the License.

DockerHost = localhost:32000

build-lightpeer:
	docker build -t $(DockerHost)/lightpeer:v0.1.0-k8s -f ../../src/lightpeer/main/Dockerfile ../../src/lightpeer/
	docker push $(DockerHost)/lightpeer:v0.1.0-k8s

deploy-controller:
	docker build -t $(DockerHost)/klight.controller:v0.1.0-k8s ../../src/klight/controller
	docker push $(DockerHost)/klight.controller:v0.1.0-k8s

	k apply -f ../../src/klight/controller/k8s/controller.yaml

deploy-example:
	docker build -t $(DockerHost)/persister:0.1.0 ./persister
	docker push $(DockerHost)/persister:0.1.0
	kubectl apply -f persister/

	docker build -t $(DockerHost)/reader:0.1.0 ./reader
	docker push $(DockerHost)/reader:0.1.0
	kubectl apply -f reader/
