package graph_test

import (
	"context"
	"fmt"
	"testing"
)

// metricHour0 is an hour boundary; metricHour1 the next hour, same UTC day.
const (
	metricHour0 = int64(1699999200)
	metricHour1 = metricHour0 + 3600
)

// TestFunctionMetricsQuery checks that Function.metrics returns server-computed
// derived values (error rate, average duration) rather than raw sums, and that
// it folds buckets into the window summary.
func TestFunctionMetricsQuery(t *testing.T) {
	c, db, _, _ := newTestClient(t)
	seedFunction(t, db, "fn1", "hello", "return 1")

	ctx := context.Background()
	// Hour 0: two executions, one an error (durations 100ms + 300ms).
	if err := db.IncrementMetricBucket(ctx, "fn1", metricHour0, false, 100); err != nil {
		t.Fatalf("IncrementMetricBucket: %v", err)
	}
	if err := db.IncrementMetricBucket(ctx, "fn1", metricHour0, true, 300); err != nil {
		t.Fatalf("IncrementMetricBucket: %v", err)
	}

	var resp struct {
		Function struct {
			Metrics struct {
				Summary struct {
					Count         int
					ErrorCount    int
					ErrorRate     float64
					AvgDurationMs float64
					MaxDurationMs int
				}
				Buckets []struct {
					BucketStart   int
					Count         int
					ErrorCount    int
					AvgDurationMs float64
					MaxDurationMs int
				}
				Granularity string
			}
		}
	}

	c.MustPost(fmt.Sprintf(`{
		function(id: "fn1") {
			metrics(from: %d, to: %d, granularity: hour) {
				summary { count errorCount errorRate avgDurationMs maxDurationMs }
				buckets { bucketStart count errorCount avgDurationMs maxDurationMs }
				granularity
			}
		}
	}`, metricHour0, metricHour1), &resp)

	s := resp.Function.Metrics.Summary
	if s.Count != 2 || s.ErrorCount != 1 {
		t.Errorf("summary counts = {count:%d err:%d}, want {2 1}", s.Count, s.ErrorCount)
	}
	if s.ErrorRate != 0.5 {
		t.Errorf("errorRate = %v, want 0.5", s.ErrorRate)
	}
	if s.AvgDurationMs != 200 {
		t.Errorf("avgDurationMs = %v, want 200", s.AvgDurationMs)
	}
	if s.MaxDurationMs != 300 {
		t.Errorf("maxDurationMs = %v, want 300", s.MaxDurationMs)
	}
	if resp.Function.Metrics.Granularity != "hour" {
		t.Errorf("granularity = %q, want hour", resp.Function.Metrics.Granularity)
	}
	if n := len(resp.Function.Metrics.Buckets); n != 1 {
		t.Fatalf("buckets = %d, want 1", n)
	}
	if b := resp.Function.Metrics.Buckets[0]; b.AvgDurationMs != 200 || b.Count != 2 {
		t.Errorf("bucket[0] = %+v, want count:2 avg:200", b)
	}
}

// TestFunctionMetricsEmpty checks the zero-data path: derived values are zero,
// not NaN, when there are no executions.
func TestFunctionMetricsEmpty(t *testing.T) {
	c, db, _, _ := newTestClient(t)
	seedFunction(t, db, "fn1", "hello", "return 1")

	var resp struct {
		Function struct {
			Metrics struct {
				Summary struct {
					Count         int
					ErrorRate     float64
					AvgDurationMs float64
				}
				Buckets []struct{ Count int }
			}
		}
	}

	c.MustPost(fmt.Sprintf(`{
		function(id: "fn1") {
			metrics(from: %d, to: %d) {
				summary { count errorRate avgDurationMs }
				buckets { count }
			}
		}
	}`, metricHour0, metricHour1), &resp)

	s := resp.Function.Metrics.Summary
	if s.Count != 0 || s.ErrorRate != 0 || s.AvgDurationMs != 0 {
		t.Errorf("empty summary = %+v, want all zero", s)
	}
	if len(resp.Function.Metrics.Buckets) != 0 {
		t.Errorf("buckets = %d, want 0", len(resp.Function.Metrics.Buckets))
	}
}
