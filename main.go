package main

import (
        "k8s.io/client-go/kubernetes"
        "log"
        "time"
        "k8s.io/client-go/rest"
        v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        v1beta1 "k8s.io/api/extensions/v1beta1"
        extv1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
        conf "github.com/ashiddo11/k8s-custom-hpa/config"
        "github.com/ashiddo11/k8s-custom-hpa/util"
        "math"
)

func scale(client extv1.DeploymentInterface, deployment string, deploymentConfig *v1beta1.Scale) (result *v1beta1.Scale) {

	result, err := client.UpdateScale(deployment, deploymentConfig)
	if err != nil {
		log.Printf(err.Error())
	}

	if err == nil {
		log.Printf("Scaled %s to %d replicas" , deployment ,result.Spec.Replicas)
	}
	return
}

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
                deploymentsClient := clientset.ExtensionsV1beta1().Deployments("default")
                for _, app := range appConfig.Apps {
                        log.Printf("Checking %s", app.Name)
                        deploymentScale, err := deploymentsClient.GetScale(app.Name, v1.GetOptions{})
                        if err != nil {
                                log.Println("[WARNING]", err)
                                continue
                        }
                        triggered := util.CheckMetric(app.Query)
                        switch {
                        case triggered && deploymentScale.Spec.Replicas < app.MaxReplicas:
                                log.Printf("Scaling up %s", app.Name)
                                deploymentScale.Spec.Replicas = int32(math.Min(float64(app.MaxReplicas), float64(deploymentScale.Spec.Replicas + app.ScaleFactor)))
                                scale(deploymentsClient, app.Name, deploymentScale)
                        case !triggered && deploymentScale.Spec.Replicas > app.MinReplicas:
                                log.Printf("Scaling down %s", app.Name)
                                deploymentScale.Spec.Replicas = int32(math.Max(float64(app.MinReplicas), float64(deploymentScale.Spec.Replicas - app.ScaleFactor)))
                                scale(deploymentsClient, app.Name, deploymentScale)
                        case triggered && deploymentScale.Spec.Replicas >= app.MaxReplicas:
                                log.Printf("Reached maximum replicas, can't scale up anymore")
                        default:
                                log.Printf("No need to scale")
                        }
                }
                time.Sleep(120 * time.Second)

        }
}
