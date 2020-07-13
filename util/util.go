package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PaesslerAG/gval"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type queryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Name      string `json:"__name__"`
				Endpoint  string `json:"endpoint"`
				Instance  string `json:"instance"`
				Job       string `json:"job"`
				Namespace string `json:"namespace"`
				Pod       string `json:"pod"`
				Service   string `json:"service"`
			} `json:"metric"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

var ctx context.Context

func init() {
	ctx = context.TODO()
}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Printf("Cannot find EnvVar %s, falling back to %s", key, fallback)
	return fallback
}

func getNamespaces(client *kubernetes.Clientset) (vpcs []string) {
	namespaces, _ := client.CoreV1().Namespaces().List(ctx, v1.ListOptions{})
	for _, ns := range namespaces.Items {
		vpcs = append(vpcs, ns.ObjectMeta.Name)
	}
	return
}

func GetDeployment(client *kubernetes.Clientset, deployment string) (deploymentScale *v1beta1.Scale, vpcFound string) {
	vpcs := getNamespaces(client)
	for _, vpc := range vpcs {
		deploymentsClient := client.ExtensionsV1beta1().Deployments(vpc)
		scale, err := deploymentsClient.GetScale(ctx, deployment, v1.GetOptions{})
		if err == nil {
			vpcFound = vpc
			log.Printf("Found %s in %s namespace", deployment, vpcFound)
			deploymentScale = scale
			break
		}
	}
	return
}

func ScaleDeployment(client *kubernetes.Clientset, vpc string, deployment string, deploymentConfig *v1beta1.Scale) (result *v1beta1.Scale) {
	deploymentsClient := client.ExtensionsV1beta1().Deployments(vpc)
	result, err := deploymentsClient.UpdateScale(ctx, deployment, deploymentConfig, v1.UpdateOptions{})
	if err != nil {
		log.Printf(err.Error())
	}

	if err == nil {
		log.Printf("Scaled %s to %d replicas", deployment, result.Spec.Replicas)
	}
	return
}

func apiCall(query string) (response *queryResponse) {
	promEP := GetEnv("PROM_ENDPOINT", "prometheus:9090")
	url := "http://" + promEP + "/api/v1/query"

	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Content-Type", "application/json")
	q := req.URL.Query()
	q.Add("query", query)
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		log.Printf("error %s", err)
		return nil
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Println(readErr)
	}

	jsonErr := json.Unmarshal(body, &response)
	if jsonErr != nil {
		log.Println(jsonErr)
	}

	if response.Status != "success" {
		log.Printf("Query unsuccessful %+v", req.URL)
	}

	return
}

func CheckMetric(query string, stackdriver bool) (triggered bool) {
	triggered = false
	if stackdriver {
		value, err := readTimeSeriesValue(query)
		if err != nil {
			log.Println(err)
			triggered = false
		}
		condition := strings.ReplaceAll(strings.Split(query, "condition=")[1], "\"", "")
		_, err = gval.Evaluate("value "+condition, map[string]interface{}{
			"value": value,
		})
		if err != nil {
			fmt.Println(err)
			triggered = false
			return
		}
		triggered = true
	} else {
		response := apiCall(query)
		if response != nil {
			if len(response.Data.Result) == 0 {
				log.Printf("Query not matched")
				triggered = false
			} else {
				log.Printf("Metric found")
				triggered = true
			}
		}
	}
	return
}
