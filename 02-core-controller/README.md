# Custom Controller - Core Resource Handling

## What is this?

An example of a custom controller that's only purpose is to delete the completed job whose completed time is large than retention time (in all namespace). We also apply learning 01's leader election to support high available. For more knowledge about how to create custom controller, please refer to Extending Kubernetes: [Create Controllers for Core and Custom Resources](https://medium.com/@trstringer/create-kubernetes-controllers-for-core-and-custom-resources-62fc35ad64a3)

In order to implement a custom controller, there's five steps to do:
* Construct a skeleton of base controller to handle resource 
* Create a `JobController` to support job retention as demo
* Implement the main func to provide flags as arguments, initialise K8s Clientset, K8s Informer, Job Controller
* Use leader election package for high available
* Define the deployment, service account and its RBAC role and cluster roles.

## Running

```
# install controller as kubernetes deployment
$ kubectl apply -f deploy/install.yaml
# create some job resources to trigger event
$ kubectl create -f deploy/example.yaml
```

## Link

- [A deep dive into Kubernetes controllers](https://engineering.bitnami.com/articles/a-deep-dive-into-kubernetes-controllers.html)
- [Create Controllers for Core and Custom Resources](https://medium.com/@trstringer/create-kubernetes-controllers-for-core-and-custom-resources-62fc35ad64a3)
- [Repository for sample controller](https://github.com/kubernetes/sample-controller/blob/master/controller.go)

