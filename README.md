# k8s-dummy-operator

A simple demo operator, that manages resources of type `Dummy`, creates matching Pods and reflects their status to the
owning `Dummy`. Based on [Operator SDK](https://sdk.operatorframework.io/).

## Getting Started

1. `git clone` this repository
2. Connect to a Kubernetes cluster (for example [minikube](https://minikube.sigs.k8s.io))
3. Install the operator to your cluster: `make deploy IMG="ghcr.io/m-hofmann/k8s-dummy-operator:v0.0.1"`
4. Create two new sample `Dummy` objects by using `kubectl apply -f demo/just_dummies.yaml`
5. Watch the status of 
   - your newly created `Dummy` objects: `kubectl describe Dummy`
   - the Pods: `kubectl get pods`

If you want to remove the operator, use `make undeploy`

## Development

For developing the operator, you can use the following commands - prerequisite: You're logged into a k8s cluster!

```shell
# if you've changed a *_types.go file
make generate
make manifests
# for running the operator on the host
make install run
```

Once you want to release a version, use this command. It targets [GitHub container registry](https://github.com/features/packages)
but can be adjusted according to your needs:

```shell
make docker-build docker-push IMG="ghcr.io/<YOUR-GITHUB-ACCOUNT-NAME>/k8s-dummy-operator"
```

## TODO

Some ideas for improving this operator:

1. Make created dependent objects more resilient
   1. Add Liveness/Readiness probe to created Pods
   2. Use Deployments instead of Pods to make `Dummy` scale out
2. Extend `Dummy` to reflect detailed status messages from Pod (e.g. `ImagePullBackoff`)
3. Make Controller heal Pod specs if they are changed
4. Implement some tests (and a lot of mocks)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

