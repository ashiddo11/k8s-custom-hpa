# k8s-custom-hpa
Kubernetes HPA using prometheus queries

## Configuration
* Yaml config file used to scale applications, mounted at /config/config.yaml

  ```apps:
    - name: hello-world
      maxReplicas: 5
      minReplicas: 1
      query: up
      scaleFactor: 2```
    
## Environment Variables

| EnvVar  | Description | Default |
| ------  | ------ | -------------|
| PROM_ENDPOINT | Prometheus endoint to query | prometheus:9090 |
| CHECK_INTERVAL | how often to query prometheus (in seconds) | 120 |


## Build
`docker build . -t k8s-custom-hpa`

## Run the autoscaler
Needs to be run within a kubernetes cluster
`helm install -n autoscaler helm/k8s-custom-hpa`


