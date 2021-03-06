rkt-compose
===========

A tool to quickly prepare and run rkt pods

## Features
* Write simplified pod-templates in yaml
* Automatic fetching of images
* ACI and Docker URLs supported
* specify networks
* creates appc conform pod-manifests
* start/stop/restart/status commands
* log viewing of your pod

## Example Template
```yaml
# Example Pod Template
# comments are allowed, thanks yaml ;)
# This defines a pod of two apps: etcd and debian
# The debian app is only for illustrative purpuses, but can be a good idea
# to include if you want to debug your pod at runtime.
---
name: etcd-example
# you can specify cpu and memory isolators!
cpu: 250m
memory: 32M
# networks to join can be specified here
networks:
  - my-net
manifest: # This maps one to one to the pod-manifest.
  apps:
    - name: etcd
      image:
        # No ID needs to be specified here.
        name: quay.io/coreos/etcd
        labels:
          - name: version
            value: v3.2.0
      app:
        exec: [ /usr/local/bin/etcd, --log-output, stdout ]
        # user and group are defaulting to "0"
    - name: debian
      image:
        name: docker://debian:testing # docker url support!
      app:
        exec: [ tail, -f, /dev/null ]
```

## Quickstart
1. Install rkt-compose: `go get github.com/trusch/rkt-compose`
2. Make it available for all users: `sudo ln -s $GOPATH/bin/rkt-compose /usr/local/bin/rkt-compose`
3. Create a folder and paste the example template into a file named `rkt-compose.yaml`
4. Run `sudo rkt-compose run -i`
5. See your pod starting up
6. Press `Ctrl-C` to stop the pod (you are in interactive mode via the `-i` flag)
7. Run `sudo rkt-compose start` to start your pod in the background
8. Check your pod with `sudo rkt-compose status`
9. Check the logs of etcd: `sudo rkt-compose logs -- -u etcd`
10. Stop our pod with `sudo rkt-compose stop`
