package util

import (
        "encoding/json"
        "k8s.io/client-go/kubernetes"
        "io/ioutil"
        "log"
        "net/http"
        "os"
        v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "time"
        v1beta1 "k8s.io/api/extensions/v1beta1"
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

func GetEnv(key, fallback string) string {
        if value, ok := os.LookupEnv(key); ok {
                return value
        }
        log.Printf("Cannot find EnvVar %s, falling back to %s", key, fallback)
        return fallback
}

func getNamespaces(client *kubernetes.Clientset) (vpcs []string) {
        namespaces, _ := client.Core().Namespaces().List(v1.ListOptions{})
        for _, ns := range namespaces.Items {
                vpcs = append(vpcs, ns.ObjectMeta.Name)
        }
        return
}

func GetDeployment(client *kubernetes.Clientset, deployment string) (deploymentScale *v1beta1.Scale, vpcFound string) {
        vpcs := getNamespaces(client)
        for _, vpc := range vpcs {
                deploymentsClient := client.ExtensionsV1beta1().Deployments(vpc)
                scale, err := deploymentsClient.GetScale(deployment, v1.GetOptions{})
                if err != nil {
                        log.Printf("%+v", err)
                } else {
                        vpcFound = vpc
                        deploymentScale = scale
                        break
                }
        }
        return
}

func ScaleDeployment(client *kubernetes.Clientset, vpc string, deployment string, deploymentConfig *v1beta1.Scale) (result *v1beta1.Scale) {
        deploymentsClient := client.ExtensionsV1beta1().Deployments(vpc)
        result, err := deploymentsClient.UpdateScale(deployment, deploymentConfig)
        if err != nil {
                log.Printf(err.Error())
        }

        if err == nil {
                log.Printf("Scaled %s to %d replicas" , deployment ,result.Spec.Replicas)
        }
        return
}

func apiCall(query string) (response *queryResponse) {
        promEP := GetEnv("PROM_ENDPOINT", "192.168.99.100:30900")
        log.Printf(promEP)
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

func CheckMetric(query string) (triggered bool) {
        response := apiCall(query)
        if len(response.Data.Result) == 0 {
                log.Printf("Query not matched")
                triggered = false
        } else {
                log.Printf("Metric found")
                triggered = true
        }
        return
}
