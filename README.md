Following the guide at https://minikube.sigs.k8s.io/docs/start/

>
> `minikube` provisions and manages local Kubernetes clusters optimized for development workflows.
>
> `kubectl` controls the Kubernetes cluster manager.
>

## Configuration

By default, both tools are looking at my local setup. Why though?

`kubectl` [has a context](https://kubernetes.io/docs/reference/kubectl/cheatsheet/#kubectl-context-and-configuration). Right now my context is `minikube`.

```sh
$ kubectl config view
apiVersion: v1
clusters:
- cluster:
    certificate-authority: ~/.minikube/ca.crt
    extensions:
    - extension:
        last-update: Wed, 22 Mar 2023 10:46:43 EDT
        provider: minikube.sigs.k8s.io
        version: v1.29.0
      name: cluster_info
    server: https://127.0.0.1:59852
  name: minikube
contexts:
- context:
    cluster: minikube
    extensions:
    - extension:
        last-update: Wed, 22 Mar 2023 10:46:43 EDT
        provider: minikube.sigs.k8s.io
        version: v1.29.0
      name: context_info
    namespace: default
    user: minikube
  name: minikube
current-context: minikube
...
```

That configuration actually lives at: `~/.kube/config`

## Removing Things

So I've gone through the hello-minikube setup and now want to run the more complex ingress example without clogging up my minikube setup with old tutorials.

> What's the opposite of `kubectl apply` if I want to do something different?

[via Stackoverflow](https://stackoverflow.com/questions/57683206/what-is-the-reverse-of-kubectl-apply)

To deploy the echo-server, I did the following:

```sh
$ kubectl create deployment hello-minikube --image=kicbase/echo-server:1.0
$ kubectl expose deployment hello-minikube --type=NodePort --port=8080
```

And I see the following:

```sh
$ kubectl get services
NAME             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
hello-minikube   NodePort    10.105.143.51   <none>        8080:32290/TCP   6s
kubernetes       ClusterIP   10.96.0.1       <none>        443/TCP          35m

$ kubectl get deployments
NAME             READY   UP-TO-DATE   AVAILABLE   AGE
hello-minikube   1/1     1            1           77s

$ kubectl get pods
NAME                              READY   STATUS    RESTARTS   AGE
hello-minikube-77b6f68484-r8vrc   1/1     Running   0          51s
```

I can delete the pod, but there's still a service and deployment, so a replacement pod is created and started. This is playing whack-a-mole, but shows the system is working.

So it makes sense that I could delete the deployment to remove the `hello-minikube` pod.

```sh
$ kubectl delete deployment hello-minikube
deployment.apps "hello-minikube" deleted

$ kubectl get pods
No resources found in default namespace.

$ kubectl get services
NAME             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
hello-minikube   NodePort    10.105.143.51   <none>        8080:32290/TCP   9m39s
kubernetes       ClusterIP   10.96.0.1       <none>        443/TCP          44m
```

Hmm, deleting the service still needed.

```sh
$ kubectl delete service hello-minikube
service "hello-minikube" deleted

$ kubectl get services
NAME         TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   49m
```

## Creating A Deployment

```sh
$ kubectl apply -f deployment.yaml
deployment.apps/hello-deployment created

$ kubectl get pods
NAME                                READY   STATUS    RESTARTS   AGE
hello-deployment-5cf4dcb5c9-cmflq   1/1     Running   0          6s
hello-deployment-5cf4dcb5c9-hz6zc   1/1     Running   0          6s
hello-deployment-5cf4dcb5c9-xjc64   1/1     Running   0          6s
```

Getting to the pods, we'd need a load balancer. Will this work?

```sh
$ kubectl expose deployment hello-deployment --type=LoadBalancer --port=8080
```

Don't know? Try the `minikube tunnel` command.

```sh
$ minikube tunnel
✅  Tunnel successfully started

📌  NOTE: Please do not close this terminal as this process must stay alive for the tunnel to be accessible ...

🏃  Starting tunnel for service hello-deployment.

$ curl localhost:8080
Request served by hello-deployment-5cf4dcb5c9-cmflq

HTTP/1.1 GET /

Host: localhost:8080
Accept: */*
User-Agent: curl/7.79.1
```

Hey! See that pod name! `hello-deployment-5cf4dcb5c9-cmflq` says our pods are getting traffic. How can we see logs from one pod?

```sh
$ kubectl exec -it pod/hello-deployment-5cf4dcb5c9-hz6zc -- ls
OCI runtime exec failed: exec failed: unable to start container process: exec: "ls": executable file not found in $PATH: unknown
command terminated with exit code 126
```

Ha! Nope, the image was created with nothing in it except the executable.

We can [see the commands that created the image](https://hub.docker.com/layers/kicbase/echo-server/1.0/images/sha256-a82eba7887a40ecae558433f34225b2611dc77f982ce05b1ddb9b282b780fc86?context=explore)--which is not the same as the Dockerfile, which lives on someone else's machine--on docker hub:

```Dockerfile
ARG TARGETPLATFORM
COPY artifacts/build/release/linux/amd64/echo-server /bin/echo-server # buildkit
ENV PORT=8080
EXPOSE map[8080/tcp:{}]
ENTRYPOINT ["/bin/echo-server"]
```

## Making a deployment + service

```sh
$ cd echo
$ make build
$ make build-image
docker build -t abachman/echo-local .
 => ...
 => => naming to docker.io/abachman/echo-local # <---- we care about this name

$ kubectl get deployments
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
hello-deployment   3/3     3            3           29ho

$ kubectl create deployment my-deployment --image=docker.io/abachman/echo-local  # <---- see
deployment.apps/my-deployment created

$ kubectl get deployments
NAME               READY   UP-TO-DATE   AVAILABLE   AGE
hello-deployment   3/3     3            3           29h
my-deployment      0/1     1            0           6m46s

$ kubectl get services
NAME               TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
hello-deployment   LoadBalancer   10.102.109.114   127.0.0.1     8080:30500/TCP   29h
kubernetes         ClusterIP      10.96.0.1        <none>        443/TCP          30h

$ kubectl expose deployment my-deployment --type=NodePort --port=9999
service/my-deployment exposed

$ kubectl get services
NAME               TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)          AGE
hello-deployment   LoadBalancer   10.102.109.114   127.0.0.1     8080:30500/TCP   29h
kubernetes         ClusterIP      10.96.0.1        <none>        443/TCP          30h
my-deployment      NodePort       10.99.161.42     <none>        9999:30442/TCP   3s
```

Hmmmmmmm. The deployment exists, service is trying to start, but there's an error with the pod because k8s can't find the image.

Delete everything, check out a guide: https://medium.com/swlh/how-to-run-locally-built-docker-images-in-kubernetes-b28fbc32cc1d

Setup `my-deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-deployment
  labels:
    app: local-echo
spec:
  replicas: 2
  selector:
    matchLabels:
      app: local-echo
  template:
    metadata:
      labels:
        app: local-echo
    spec:
      containers:
      - name: local-echo
        image: abachman/echo-local:latest
        imagePullPolicy: Never # <---- this is key
        ports:
        - containerPort: 9999
```

> To fix this, I use the minikube docker-env command that outputs environment variables needed to point the local Docker daemon to the minikube internal Docker registry:

```sh
$ minikube docker-env
export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://127.0.0.1:59854"
export DOCKER_CERT_PATH="/Users/abachman/.minikube/certs"
export MINIKUBE_ACTIVE_DOCKERD="minikube"

# To point your shell to minikube's docker-daemon, run:
# eval $(minikube -p minikube docker-env)

$ eval $(minikube -p minikube docker-env) # <-- run the recommended command
```

> I now need to build the image once again, so that it’s installed in the minikube registry, instead of the local one.

```sh
$ docker build echo/ -t abachman/echo-local
Sending build context to Docker daemon  15.24MB
Step 1/4 : FROM scratch
 --->
Step 2/4 : COPY build/echo-linux /bin/echo
 ---> 43bf0be6ba5e
Step 3/4 : CMD ["/bin/echo"]
 ---> Running in c71ca8ab20cb
Removing intermediate container c71ca8ab20cb
 ---> 0d8e6348fa99
Step 4/4 : EXPOSE 9999
 ---> Running in 44ca0cb7b30d
Removing intermediate container 44ca0cb7b30d
 ---> aca7a9db0ff8
Successfully built aca7a9db0ff8
Successfully tagged abachman/echo-local:latest
```

Now create with kubectl:

```sh
$ kubectl create -f my-deployment
```

## Upgrading Image

I made a change to `echo/server.go`, now I want to build a new image and push it to my deployment.

```sh
$ docker build echo/ -t abachman/echo-local:0.0.2
Step 1/4:
...
Successfully built 9ea4f90d1341
Successfully tagged abachman/echo-local:0.0.2
```

update my-deployment.yaml:

```yaml
        image: abachman/echo-local:latest
        # becomes
        image: abachman/echo-local:0.0.2
```

and apply config:

```sh
$ kubectl apply -f my-deployment.yaml
Warning: resource deployments/my-deployment is missing the kubectl.kubernetes.io/last-applied-configuration annotation which is required by kubectl apply. kubectl apply should only be used on resources created declaratively by either kubectl create --save-config or kubectl apply. The missing annotation will be patched automatically.
deployment.apps/my-deployment configured

$ minikube service my-deployment
|-----------|---------------|-------------|---------------------------|
| NAMESPACE |     NAME      | TARGET PORT |            URL            |
|-----------|---------------|-------------|---------------------------|
| default   | my-deployment |        9999 | http://192.168.49.2:30258 |
|-----------|---------------|-------------|---------------------------|
🏃  Starting tunnel for service my-deployment.
|-----------|---------------|-------------|------------------------|
| NAMESPACE |     NAME      | TARGET PORT |          URL           |
|-----------|---------------|-------------|------------------------|
| default   | my-deployment |             | http://127.0.0.1:59076 |
|-----------|---------------|-------------|------------------------|
🎉  Opening service default/my-deployment in default browser...
❗  Because you are using a Docker driver on darwin, the terminal needs to be open to run it.
```

Now from a new console:

```sh
$ curl 'http://localhost:59076/path?q=something'
response from my-deployment-7744767dc8-2dfb6
GET /path?q=something HTTP/1.1
Host: localhost:59076
User-Agent: curl/7.79.1
Accept: */*
```

It does the thing.
