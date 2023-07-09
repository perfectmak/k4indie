# K4Indie 👷🏼‍♀️☸️ 

K4Indie _(Kay-for-Indie_)_ is a Kubernetes operator to simplify the deployment of applications to Kubernetes. 

> 🚧🚧 This is not production ready and is still heavily a work in progress. Check the open issues for WIP and Roadmap 🚧🚧 

## Description
K4Indie is focused on bringing a simple Heroku-like experience to deploy and run applications. This makes it a perfect candidate to run your application on cheap cloud-hosted Kubernetes platforms like DigitalOcean.

It exports a Custom Resource named `Application` that can be used to define an application's requirement.
After the Application resource is applied to the Kubernetes cluster, K4Indie will set up all the required native resources to ensure your application is running and accessible to the internet (in cases where that is necessary).

## Getting Started
First, install this operator and its custom resource definitions in your cluster:

> TBD: It is yet to be packaged for re-distribution, but you can the instructions on the 'Deploying to the Cluster' for how to deploy from this repository


Next, to make your application accessible to the internet, you'll need
to install an Ingress controller. An easy one to install is [nginx-ingress](https://docs.nginx.com/nginx-ingress-controller/installation/installation-with-helm/).

> The step to install ingress controller will not be required in upcoming releases because it will be bundled as part of K4Indie.

The Kubernetes cluster is not ready to deploy applications. There are sample application definitions in the `config/samples`. Try deploying the nginx sample application by running the following command:

```
kubectl apply -f config/samples/operators_v1alpha1_nginx.yaml
```

Visit port `8080` of the cluster's IP address to preview the application.

## `Application` Specification
The major resource for managing application workloads is the `Application` resource. 

> A proper specification document coming up soon. In the meantime, the OpenAPI Schema can be found [here](config/crd/bases/operators.k4indie.io_applications.yaml). Also, explore the `config/samples` directory for some example Application definitions.

## Contributing
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.

**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

Before making a contribution, please open an issue to discuss your plans and get feedback.

### Deploying to the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/operator:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/operator:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller from the cluster:

```sh
make undeploy
```

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

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

