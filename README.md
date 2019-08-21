# kafka-provider

(work-in-progress)

deploy the kafka-provider:

```
GO111MODULE=on ko apply -f config/
```

forward localhost port 8080 to the deployment:

```
kubectl port-forward deployment/kafka-provider 8080:8080
```

create a topic:

```
curl -X PUT localhost:8080/providertest1
```

