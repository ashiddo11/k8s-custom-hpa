package main

import (
        "k8s.io/client-go/kubernetes"
        "log"
        "time"
        "k8s.io/client-go/rest"
        conf "github.com/ashiddo11/k8s-custom-hpa/config"
        "github.com/ashiddo11/k8s-custom-hpa/util"
        "strconv"
        "math"
)

func main() {

        config, err := rest.InClusterConfig()
        if err != nil {
             panic(err.Error())
        }

        // create the clientset
        clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
                panic(err.Error())
        }

        for {
                appConfig, _ := conf.LoadConfig()
                for _, app := range appConfig.Apps {
                        log.Printf("Checking %s", app.Name)
                        deploymentScale, ns := util.GetDeployment(clientset, app.Name)
                        if deploymentScale == nil {
                            break
                        }
                        triggered := util.CheckMetric(app.Query)
                        switch {
                        case triggered && deploymentScale.Spec.Replicas < app.MaxReplicas:
                                log.Printf("Scaling up %s", app.Name)
                                deploymentScale.Spec.Replicas = int32(math.Min(float64(app.MaxReplicas), float64(deploymentScale.Spec.Replicas + app.ScaleFactor)))
                                util.ScaleDeployment(clientset, ns, app.Name, deploymentScale)
                        case !triggered && deploymentScale.Spec.Replicas > app.MinReplicas:
                                log.Printf("Scaling down %s", app.Name)
                                deploymentScale.Spec.Replicas = int32(math.Max(float64(app.MinReplicas), float64(deploymentScale.Spec.Replicas - app.ScaleFactor)))
                                util.ScaleDeployment(clientset, ns, app.Name, deploymentScale)
                        case triggered && deploymentScale.Spec.Replicas >= app.MaxReplicas:
                                log.Printf("Reached maximum replicas, can't scale up anymore")
                        default:
                                log.Printf("No need to scale")
                        }
                }
                checkInterval, _ := strconv.Atoi(util.GetEnv("CHECK_INTERVAL", "120"))
                interval := time.Duration(checkInterval)
                time.Sleep(interval * time.Second)
        }
}
