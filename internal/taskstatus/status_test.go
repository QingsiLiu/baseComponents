package taskstatus

import "testing"

func TestName(t *testing.T) {
	cases := []struct {
		status int32
		want   string
	}{
		{status: Pending, want: "pending"},
		{status: Running, want: "running"},
		{status: Completed, want: "completed"},
		{status: Canceled, want: "canceled"},
		{status: Failed, want: "failed"},
		{status: 99, want: "unknown"},
	}

	for _, tc := range cases {
		if got := Name(tc.status); got != tc.want {
			t.Fatalf("expected status %d -> %s, got %s", tc.status, tc.want, got)
		}
	}
}
