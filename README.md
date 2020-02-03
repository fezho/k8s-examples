# Examples of programming on Kubernetes
This repositry contains a number of examples of how to program on Kubernetes. These applications are all building on Kubernetes' APIs.

It includes following subjects:
 * [x] Leader Election
 * [x] Custom Controller
 * [ ] Custom Scheduler
 * [ ] CRD and Operator
 * [ ] CNI Plugin

## Develop and Deploy

### Develop Environment
* Kubernetes 1.16
* Golang 1.12

### Project Layout
We use the basic layout for Go application projects including `Go module`, it looks like:

```
.
├── Dockerfile
├── Makefile
├── README.md
├── cmd        # main application for this project
├── deploy     # kubernetes deploy files, such as Deployment, Service, Role
│   └── install.yaml
├── go.mod     # go module file
├── go.sum     # go module file
└── pkg
```

### Publish docker image to github package
```console
# TOKEN can be generated in https://github.com/settings/tokens
docker login -u USERNAME -p TOKEN docker.pkg.github.com
make image
docker push docker.pkg.github.com/fezho/k8s-examples/kube-xxx-demo:tag
```


## License

MIT
