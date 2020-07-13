package util

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

var (
	projectID = os.Getenv("GOOGLE_PROJECT_ID")
)

func readTimeSeriesValue(query string) (value int64, err error) {
	ctx := context.Background()
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return 0, err
	}
	startTime := time.Now().UTC().Add(time.Second * -90).Unix()
	endTime := time.Now().UTC().Unix()

	//Extract query
	q := strings.Split(query, "condition=")[0]

	req := &monitoringpb.ListTimeSeriesRequest{
		Name:   "projects/" + projectID,
		Filter: q,
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamp.Timestamp{Seconds: startTime},
			EndTime:   &timestamp.Timestamp{Seconds: endTime},
		},
	}
	iter := c.ListTimeSeries(ctx, req)

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("could not read time series value, %v ", err)
		}
		for _, point := range resp.Points {
			value = point.GetValue().GetInt64Value()
		}
	}

	return value, nil
}
