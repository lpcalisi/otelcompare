package trace

import (
	"strings"
	"testing"
	"time"
)

func TestParseTraces(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "valid traces",
			input:   []byte(`[{"trace_id": "trace1", "spans": [{"span_id": "span1", "name": "test", "start_time": "2024-03-07T00:00:00Z", "end_time": "2024-03-07T00:00:01Z"}]}]`),
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   []byte(`invalid json`),
			wantErr: true,
		},
		{
			name:    "empty array",
			input:   []byte(`[]`),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTraces(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTraces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(got) == 0 && tt.name != "empty array" {
				t.Error("ParseTraces() returned empty slice for valid input")
			}
		})
	}
}

func TestGetTraceIdentifier(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		trace     Trace
		attribute string
		expected  string
	}{
		{
			name: "root span exists",
			trace: Trace{
				Spans: []Span{
					{SpanID: "span1", ParentSpanID: "", Name: "root", StartTime: now, EndTime: now.Add(time.Second)},
					{SpanID: "span2", ParentSpanID: "span1", Name: "child", StartTime: now, EndTime: now.Add(time.Second)},
				},
			},
			attribute: "name",
			expected:  "root",
		},
		{
			name: "no root span",
			trace: Trace{
				Spans: []Span{
					{SpanID: "span1", ParentSpanID: "span2", Name: "child1", StartTime: now, EndTime: now.Add(time.Second)},
					{SpanID: "span2", ParentSpanID: "span1", Name: "child2", StartTime: now, EndTime: now.Add(time.Second)},
				},
			},
			attribute: "name",
			expected:  "child1",
		},
		{
			name:      "empty spans",
			trace:     Trace{Spans: []Span{}},
			attribute: "name",
			expected:  "Unknown Trace",
		},
		{
			name: "by trace_id",
			trace: Trace{
				TraceID: "test-trace",
				Spans: []Span{
					{Name: "test-span", StartTime: now, EndTime: now.Add(time.Second)},
				},
			},
			attribute: "trace_id",
			expected:  "test-trace",
		},
		{
			name: "by attribute",
			trace: Trace{
				Attributes: map[string]string{
					"test-attr": "test-value",
				},
				Spans: []Span{
					{Name: "test-span", StartTime: now, EndTime: now.Add(time.Second)},
				},
			},
			attribute: "test-attr",
			expected:  "test-value",
		},
		{
			name: "by resource attribute",
			trace: Trace{
				ResourceAttrs: map[string]string{
					"test-attr": "test-value",
				},
				Spans: []Span{
					{Name: "test-span", StartTime: now, EndTime: now.Add(time.Second)},
				},
			},
			attribute: "test-attr",
			expected:  "test-value",
		},
		{
			name: "fallback to trace_id",
			trace: Trace{
				TraceID: "test-trace",
				Spans: []Span{
					{Name: "test-span", StartTime: now, EndTime: now.Add(time.Second)},
				},
			},
			attribute: "non-existent",
			expected:  "test-trace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTraceIdentifier(tt.trace, tt.attribute)
			if got != tt.expected {
				t.Errorf("getTraceIdentifier() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetTraceDuration(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		trace    Trace
		expected time.Duration
	}{
		{
			name: "single span",
			trace: Trace{
				Spans: []Span{
					{
						StartTime: now,
						EndTime:   now.Add(time.Second),
					},
				},
			},
			expected: time.Second,
		},
		{
			name: "multiple spans",
			trace: Trace{
				Spans: []Span{
					{
						StartTime: now,
						EndTime:   now.Add(2 * time.Second),
					},
					{
						StartTime: now.Add(time.Second),
						EndTime:   now.Add(3 * time.Second),
					},
				},
			},
			expected: 3 * time.Second,
		},
		{
			name:     "empty spans",
			trace:    Trace{Spans: []Span{}},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTraceDuration(tt.trace)
			if got != tt.expected {
				t.Errorf("getTraceDuration() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "microseconds",
			duration: 500 * time.Microsecond,
			expected: "500.00µs",
		},
		{
			name:     "milliseconds",
			duration: 500 * time.Millisecond,
			expected: "500.00ms",
		},
		{
			name:     "seconds",
			duration: 5 * time.Second,
			expected: "5.00s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.expected {
				t.Errorf("formatDuration() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTruncateID(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected string
	}{
		{
			name:     "long id",
			id:       "1234567890",
			expected: "12345678",
		},
		{
			name:     "short id",
			id:       "123",
			expected: "123",
		},
		{
			name:     "empty id",
			id:       "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateID(tt.id)
			if got != tt.expected {
				t.Errorf("truncateID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCompareTraces(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		traces1  []Trace
		traces2  []Trace
		contains []string
	}{
		{
			name: "matching traces",
			traces1: []Trace{
				{
					TraceID: "trace1",
					Spans: []Span{
						{Name: "span1", StartTime: now, EndTime: now.Add(time.Second)},
					},
				},
			},
			traces2: []Trace{
				{
					TraceID: "trace1",
					Spans: []Span{
						{Name: "span1", StartTime: now, EndTime: now.Add(2 * time.Second)},
					},
				},
			},
			contains: []string{"Matching Traces", "Duration Comparison"},
		},
		{
			name: "different traces",
			traces1: []Trace{
				{
					TraceID: "trace1",
					Spans: []Span{
						{Name: "span1", StartTime: now, EndTime: now.Add(time.Second)},
					},
				},
			},
			traces2: []Trace{
				{
					TraceID: "trace2",
					Spans: []Span{
						{Name: "span2", StartTime: now, EndTime: now.Add(time.Second)},
					},
				},
			},
			contains: []string{"Only in First File", "Only in Second File"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareTraces(tt.traces1, tt.traces2)
			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("CompareTraces() output does not contain %v", s)
				}
			}
		})
	}
}
