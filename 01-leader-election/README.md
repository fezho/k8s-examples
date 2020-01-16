# leader election
Leader election for application running on kubernetes.


## Publish docker image to github package
```console
# TOKEN can be generated in https://github.com/settings/tokens
docker login -u USERNAME -p TOKEN docker.pkg.github.com
make image
docker push ...
```


## License

MIT
