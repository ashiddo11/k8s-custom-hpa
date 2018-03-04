package util

import (
        "encoding/json"
        "io/ioutil"
        "log"
        "net/http"
        "os"
        "time"
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

func getEnv(key, fallback string) string {
        if value, ok := os.LookupEnv(key); ok {
                return value
        }
        log.Printf("Cannot find EnvVar %s, falling back to %s", key, fallback)
        return fallback
}

func apiCall(query string) (response *queryResponse) {
        promEP := getEnv("PROM_ENDPOINT", "prometheus:9090")
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
