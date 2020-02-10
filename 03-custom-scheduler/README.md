# Custom scheduler

## What is this?

This is a custom kubernetes scheduler that can be used for tutorials. It's not intended for production usage.

The scheduler watches pods and binds them to random nodes, then emits "Scheduled" events.

## Running

```
# install controller as kubernetes deployment
$ kubectl apply -f deploy/install.yaml
```

## Link
