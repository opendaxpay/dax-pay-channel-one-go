package jsonx_test

import (
	"encoding/json"
	"testing"
	"time"

	"daxpay.open/dax-pay-channel-one-go/internal/jsonx"
)

func TestInt64StringRoundTrip(t *testing.T) {
	type row struct {
		Amount jsonx.Int64String `json:"amount"`
	}
	var r row
	if err := jsonx.Unmarshal([]byte(`{"amount":"12345","extra":true}`), &r); err != nil {
		t.Fatal(err)
	}
	if r.Amount != 12345 {
		t.Fatalf("amount=%d", r.Amount)
	}
	out, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `{"amount":"12345"}` {
		t.Fatalf("marshal=%s", out)
	}
	var r2 row
	if err := jsonx.Unmarshal([]byte(`{"amount":99}`), &r2); err != nil {
		t.Fatal(err)
	}
	if r2.Amount != 99 {
		t.Fatalf("numeric amount=%d", r2.Amount)
	}
}

func TestOffsetDateTimeUTC(t *testing.T) {
	type row struct {
		T jsonx.OffsetDateTime `json:"t"`
	}
	var r row
	if err := jsonx.Unmarshal([]byte(`{"t":"2026-07-06T11:15:48Z"}`), &r); err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 7, 6, 11, 15, 48, 0, time.UTC)
	if !r.T.Time().Equal(want) {
		t.Fatalf("time=%v", r.T.Time())
	}
	out, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != `{"t":"2026-07-06T11:15:48Z"}` {
		t.Fatalf("marshal=%s", out)
	}
}
