package store

import (
	"context"
	"testing"
)

// Hour boundaries that share the same UTC day, so the daily downsample folds
// them into a single bucket.
const (
	metricHour0 = int64(1699999200) // 1700000000 truncated to the hour
	metricHour1 = metricHour0 + 3600
)

func TestSQLiteDB_MetricBuckets(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// A function is required for the execution_metrics foreign key.
	if _, err := sqliteDB.CreateFunction(ctx, Function{ID: "func_m", Name: "metrics-fn"}); err != nil {
		t.Fatalf("CreateFunction: %v", err)
	}

	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("IncrementMetricBucket: %v", err)
		}
	}
	// Two executions in hour 0 (one error), one in hour 1.
	must(sqliteDB.IncrementMetricBucket(ctx, "func_m", metricHour0, false, 100))
	must(sqliteDB.IncrementMetricBucket(ctx, "func_m", metricHour0, true, 300))
	must(sqliteDB.IncrementMetricBucket(ctx, "func_m", metricHour1, false, 50))

	// Hourly read: two distinct buckets, the first folding both executions.
	hourly, err := sqliteDB.GetFunctionMetrics(ctx, "func_m", metricHour0, metricHour1+3600, 3600)
	if err != nil {
		t.Fatalf("GetFunctionMetrics hourly: %v", err)
	}
	if len(hourly) != 2 {
		t.Fatalf("hourly buckets = %d, want 2", len(hourly))
	}
	b0 := hourly[0]
	if b0.BucketStart != metricHour0 || b0.Count != 2 || b0.ErrorCount != 1 ||
		b0.SumDurationMs != 400 || b0.MaxDurationMs != 300 {
		t.Errorf("hour0 bucket = %+v, want {start:%d count:2 err:1 sum:400 max:300}", b0, metricHour0)
	}
	if b1 := hourly[1]; b1.Count != 1 || b1.SumDurationMs != 50 || b1.MaxDurationMs != 50 {
		t.Errorf("hour1 bucket = %+v, want count:1 sum:50 max:50", b1)
	}

	// Daily downsample: both hours collapse into one day bucket.
	dayStart := (metricHour0 / 86400) * 86400
	daily, err := sqliteDB.GetFunctionMetrics(ctx, "func_m", dayStart, dayStart+86400, 86400)
	if err != nil {
		t.Fatalf("GetFunctionMetrics daily: %v", err)
	}
	if len(daily) != 1 {
		t.Fatalf("daily buckets = %d, want 1", len(daily))
	}
	d := daily[0]
	if d.BucketStart != dayStart || d.Count != 3 || d.ErrorCount != 1 ||
		d.SumDurationMs != 450 || d.MaxDurationMs != 300 {
		t.Errorf("daily bucket = %+v, want {start:%d count:3 err:1 sum:450 max:300}", d, dayStart)
	}

	// Range filter excludes buckets outside [from, to).
	none, err := sqliteDB.GetFunctionMetrics(ctx, "func_m", metricHour1+3600, metricHour1+7200, 3600)
	if err != nil {
		t.Fatalf("GetFunctionMetrics range: %v", err)
	}
	if len(none) != 0 {
		t.Errorf("out-of-range buckets = %d, want 0", len(none))
	}
}

func TestSQLiteDB_DeleteOldMetricBuckets(t *testing.T) {
	db, sqliteDB := setupTestDB(t)
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	if _, err := sqliteDB.CreateFunction(ctx, Function{ID: "func_m", Name: "metrics-fn"}); err != nil {
		t.Fatalf("CreateFunction: %v", err)
	}

	if err := sqliteDB.IncrementMetricBucket(ctx, "func_m", metricHour0, false, 100); err != nil {
		t.Fatalf("IncrementMetricBucket: %v", err)
	}
	if err := sqliteDB.IncrementMetricBucket(ctx, "func_m", metricHour1, false, 100); err != nil {
		t.Fatalf("IncrementMetricBucket: %v", err)
	}

	// Cutoff at hour 1 deletes only the strictly-older hour 0 bucket.
	deleted, err := sqliteDB.DeleteOldMetricBuckets(ctx, metricHour1)
	if err != nil {
		t.Fatalf("DeleteOldMetricBuckets: %v", err)
	}
	if deleted != 1 {
		t.Errorf("deleted = %d, want 1", deleted)
	}

	remaining, err := sqliteDB.GetFunctionMetrics(ctx, "func_m", metricHour0, metricHour1+3600, 3600)
	if err != nil {
		t.Fatalf("GetFunctionMetrics: %v", err)
	}
	if len(remaining) != 1 || remaining[0].BucketStart != metricHour1 {
		t.Errorf("remaining = %+v, want only hour1 bucket", remaining)
	}
}
