# watch-resource

**watch-resource** is a simple binary to trigger a reload when a file inside a pod changes or a Kubernetes resource (ConfigMap or Secret) is updated.

```
$ ./watch-resource-cmdrunner --help
Usage of :
      --command="": Path to the script or command to execute inside the pod
      --configmap="": Name of the configmap to watch. Takes the form namespace/name
      --container="": Name of hte container inside the pod. An empty string uses the first container in the pod
      --file="": Path to the file to watch
      --pod="": Name of the pod where the command will be executed. Takes the form namespace/name.
		No value means "this" pod (requires `POD_NAME` and `POD_NAMESPACE` using downward API)
      --secret="": Name of the secret to watch. Takes the form namespace/name
```

The flags `--file`, `--configmap` and `--secret` are mutually exclusive".


Example:
```
        /watch-resource-cmdrunner \
        --command=curl http://localhost:10254/reload-template \
        --file=/etc/nginx/template/nginx.tmpl \
        --container=nginx-ingress-lb
```
