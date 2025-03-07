package trace

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Trace represents a complete OpenTelemetry trace
type Trace struct {
	TraceID       string            `json:"trace_id"`
	Spans         []Span            `json:"spans"`
	Attributes    map[string]string `json:"attributes"`
	ResourceAttrs map[string]string `json:"resource_attributes"`
}

// Span represents a single span in a trace
type Span struct {
	SpanID       string            `json:"span_id"`
	ParentSpanID string            `json:"parent_span_id"`
	Name         string            `json:"name"`
	StartTime    time.Time         `json:"start_time"`
	EndTime      time.Time         `json:"end_time"`
	Attributes   map[string]string `json:"attributes"`
	Events       []Event           `json:"events"`
}

// Event represents an event within a span
type Event struct {
	Time       time.Time         `json:"time"`
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes"`
}

// TraceSet represents a set of traces from a single file
type TraceSet struct {
	Name   string
	Traces []Trace
}

// ParseTraces reads a JSON file and returns a slice of traces
func ParseTraces(data []byte) ([]Trace, error) {
	var traces []Trace
	if err := json.Unmarshal(data, &traces); err != nil {
		return nil, fmt.Errorf("error unmarshaling traces: %w", err)
	}
	return traces, nil
}

// GenerateMarkdown generates a Markdown representation of the traces
func GenerateMarkdown(traces []Trace) string {
	var sb strings.Builder

	// First table: Overview of traces
	sb.WriteString("**Traces Overview:**\n\n")
	sb.WriteString("| Trace ID | Trace Name | Duration | Spans |\n")
	sb.WriteString("|----------|------------|----------|-------|\n")

	// Create a map to quickly access spans by trace ID
	traceSpanMaps := make(map[string]map[string]*Span)
	for _, t := range traces {
		spanMap := make(map[string]*Span)
		for i := range t.Spans {
			spanMap[t.Spans[i].SpanID] = &t.Spans[i]
		}
		traceSpanMaps[t.TraceID] = spanMap
	}

	// Sort traces by duration (descending)
	sort.Slice(traces, func(i, j int) bool {
		iDuration := getTraceDuration(traces[i])
		jDuration := getTraceDuration(traces[j])
		return iDuration > jDuration
	})

	for _, t := range traces {
		duration := getTraceDuration(t)
		traceName := getTraceIdentifier(t, "name")
		sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %d |\n",
			truncateID(t.TraceID),
			traceName,
			formatDuration(duration),
			len(t.Spans)))
	}

	// Second table: Detailed span information
	sb.WriteString("\n**Span Details:**\n\n")
	sb.WriteString("| Trace ID | Span ID | Span Name | Duration | Parent |\n")
	sb.WriteString("|----------|---------|-----------|----------|--------|\n")

	// Sort spans by duration (descending)
	for _, t := range traces {
		spans := t.Spans
		sort.Slice(spans, func(i, j int) bool {
			return spans[i].EndTime.Sub(spans[i].StartTime) > spans[j].EndTime.Sub(spans[j].StartTime)
		})

		for _, span := range spans {
			parentName := "root"
			if span.ParentSpanID != "" {
				if parentSpan, ok := traceSpanMaps[t.TraceID][span.ParentSpanID]; ok {
					parentName = parentSpan.Name
				}
			}
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %s | %s |\n",
				truncateID(t.TraceID),
				truncateID(span.SpanID),
				span.Name,
				formatDuration(span.EndTime.Sub(span.StartTime)),
				parentName))
		}
	}

	// Expandable details for each trace
	sb.WriteString("\n**Trace Details:**\n\n")
	for _, t := range traces {
		sb.WriteString(fmt.Sprintf("<details>\n<summary>Trace %s (%s)</summary>\n\n", truncateID(t.TraceID), getTraceIdentifier(t, "name")))

		// Show trace attributes
		if len(t.Attributes) > 0 {
			sb.WriteString("**Trace Attributes:**\n\n")
			sb.WriteString("| Key | Value |\n")
			sb.WriteString("|-----|--------|\n")
			for k, v := range t.Attributes {
				sb.WriteString(fmt.Sprintf("| %s | %s |\n", k, v))
			}
			sb.WriteString("\n")
		}

		// Show spans in hierarchical order
		sb.WriteString("**Spans:**\n\n")
		showSpan(&sb, &t, "", traceSpanMaps[t.TraceID])

		sb.WriteString("</details>\n\n")
	}

	return sb.String()
}

// showSpan recursively shows a span and its children
func showSpan(sb *strings.Builder, t *Trace, parentID string, spanMap map[string]*Span) {
	// Find all spans with this parent
	for _, span := range t.Spans {
		if span.ParentSpanID == parentID {
			// Show this span
			sb.WriteString(fmt.Sprintf("- **%s** (%s)\n", span.Name, formatDuration(span.EndTime.Sub(span.StartTime))))

			// Show attributes if any
			if len(span.Attributes) > 0 {
				sb.WriteString("  **Attributes:**\n")
				for k, v := range span.Attributes {
					sb.WriteString(fmt.Sprintf("  - %s: %s\n", k, v))
				}
			}

			// Show events if any
			if len(span.Events) > 0 {
				sb.WriteString("  **Events:**\n")
				for _, event := range span.Events {
					sb.WriteString(fmt.Sprintf("  - %s\n", event.Name))
					if len(event.Attributes) > 0 {
						for k, v := range event.Attributes {
							sb.WriteString(fmt.Sprintf("    - %s: %s\n", k, v))
						}
					}
				}
			}

			// Recursively show children
			showSpan(sb, t, span.SpanID, spanMap)
		}
	}
}

// Helper functions
func truncateID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Nanoseconds())/1000.0)
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Milliseconds()))
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func getFileNameWithoutExt(fileName string) string {
	return strings.TrimSuffix(fileName, ".json")
}

func getTraceDuration(t Trace) time.Duration {
	if len(t.Spans) == 0 {
		return 0
	}

	var earliest, latest time.Time
	first := true

	for _, span := range t.Spans {
		if first {
			earliest = span.StartTime
			latest = span.EndTime
			first = false
		} else {
			if span.StartTime.Before(earliest) {
				earliest = span.StartTime
			}
			if span.EndTime.After(latest) {
				latest = span.EndTime
			}
		}
	}

	return latest.Sub(earliest)
}

// CompareTraces compares two sets of traces and generates a markdown report
func CompareTraces(traces1, traces2 []Trace) string {
	var sb strings.Builder

	// Create maps of traces by name for quick lookup
	traces1Map := make(map[string]*Trace)
	traces2Map := make(map[string]*Trace)

	for i := range traces1 {
		name := getTraceIdentifier(traces1[i], "name")
		traces1Map[name] = &traces1[i]
	}

	for i := range traces2 {
		name := getTraceIdentifier(traces2[i], "name")
		traces2Map[name] = &traces2[i]
	}

	// Compare traces
	sb.WriteString("### Trace Comparison\n\n")

	// Find matching traces
	var matchingTraces []string
	for name := range traces1Map {
		if _, exists := traces2Map[name]; exists {
			matchingTraces = append(matchingTraces, name)
		}
	}
	sort.Strings(matchingTraces)

	// Find traces only in first set
	var onlyInFirst []string
	for name := range traces1Map {
		if _, exists := traces2Map[name]; !exists {
			onlyInFirst = append(onlyInFirst, name)
		}
	}
	sort.Strings(onlyInFirst)

	// Find traces only in second set
	var onlyInSecond []string
	for name := range traces2Map {
		if _, exists := traces1Map[name]; !exists {
			onlyInSecond = append(onlyInSecond, name)
		}
	}
	sort.Strings(onlyInSecond)

	// Summary table
	sb.WriteString("**Comparison Summary:**\n\n")
	sb.WriteString("| Category | Count |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Matching Traces | %d |\n", len(matchingTraces)))
	sb.WriteString(fmt.Sprintf("| Only in First File | %d |\n", len(onlyInFirst)))
	sb.WriteString(fmt.Sprintf("| Only in Second File | %d |\n", len(onlyInSecond)))
	sb.WriteString("\n")

	// Matching traces comparison
	if len(matchingTraces) > 0 {
		sb.WriteString("**Matching Traces:**\n\n")
		for _, name := range matchingTraces {
			t1 := traces1Map[name]
			t2 := traces2Map[name]

			sb.WriteString(fmt.Sprintf("<details>\n<summary>%s</summary>\n\n", name))

			// Compare durations
			duration1 := getTraceDuration(*t1)
			duration2 := getTraceDuration(*t2)
			durationDiff := duration2 - duration1
			durationChange := (durationDiff.Seconds() / duration1.Seconds()) * 100

			sb.WriteString("**Duration Comparison:**\n\n")
			sb.WriteString("| File | Duration |\n")
			sb.WriteString("|------|----------|\n")
			sb.WriteString(fmt.Sprintf("| First | %s |\n", formatDuration(duration1)))
			sb.WriteString(fmt.Sprintf("| Second | %s |\n", formatDuration(duration2)))
			sb.WriteString(fmt.Sprintf("| Difference | %s (%.1f%%) |\n", formatDuration(durationDiff), durationChange))
			sb.WriteString("\n")

			// Compare spans
			sb.WriteString("**Span Comparison:**\n\n")
			sb.WriteString("| Span Name | First Duration | Second Duration | Difference |\n")
			sb.WriteString("|-----------|----------------|-----------------|------------|\n")

			// Create maps of spans by name
			spans1Map := make(map[string]*Span)
			spans2Map := make(map[string]*Span)

			for i := range t1.Spans {
				spans1Map[t1.Spans[i].Name] = &t1.Spans[i]
			}

			for i := range t2.Spans {
				spans2Map[t2.Spans[i].Name] = &t2.Spans[i]
			}

			// Compare matching spans
			for name, span1 := range spans1Map {
				if span2, exists := spans2Map[name]; exists {
					d1 := span1.EndTime.Sub(span1.StartTime)
					d2 := span2.EndTime.Sub(span2.StartTime)
					diff := d2 - d1
					change := (diff.Seconds() / d1.Seconds()) * 100

					sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s (%.1f%%) |\n",
						name,
						formatDuration(d1),
						formatDuration(d2),
						formatDuration(diff),
						change))
				}
			}

			sb.WriteString("\n</details>\n\n")
		}
	}

	// Traces only in first file
	if len(onlyInFirst) > 0 {
		sb.WriteString("**Traces Only in First File:**\n\n")
		for _, name := range onlyInFirst {
			sb.WriteString(fmt.Sprintf("- %s\n", name))
		}
		sb.WriteString("\n")
	}

	// Traces only in second file
	if len(onlyInSecond) > 0 {
		sb.WriteString("**Traces Only in Second File:**\n\n")
		for _, name := range onlyInSecond {
			sb.WriteString(fmt.Sprintf("- %s\n", name))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// New function to get the trace identifier based on the specified attribute
func getTraceIdentifier(t Trace, attribute string) string {
	// If the attribute is "trace_id", use the trace ID
	if attribute == "trace_id" {
		return t.TraceID
	}

	// If the attribute is "name", find the root span or first span
	if attribute == "name" {
		if len(t.Spans) == 0 {
			return "Unknown Trace"
		}

		// Try to find a root span (no parent)
		for _, span := range t.Spans {
			if span.ParentSpanID == "" {
				return span.Name
			}
		}

		// If no root span found, return the name of the first span
		return t.Spans[0].Name
	}

	// Search in trace attributes
	if value, ok := t.Attributes[attribute]; ok {
		return value
	}

	// Search in resource attributes
	if value, ok := t.ResourceAttrs[attribute]; ok {
		return value
	}

	// Fallback to trace ID
	return t.TraceID
}

// CompareMultipleTraces compares multiple sets of traces and generates a markdown report
func CompareMultipleTraces(traceSets []TraceSet, attribute string) string {
	var sb strings.Builder

	sb.WriteString("### Multiple Traces Comparison\n\n")

	// Create maps of traces by attribute for each set
	traceMaps := make([]map[string]*Trace, len(traceSets))
	for i, set := range traceSets {
		traceMaps[i] = make(map[string]*Trace)
		for j := range set.Traces {
			identifier := getTraceIdentifier(set.Traces[j], attribute)
			traceMaps[i][identifier] = &set.Traces[j]
		}
	}

	// Find all unique trace names across all sets
	allTraceNames := make(map[string]bool)
	for _, traceMap := range traceMaps {
		for name := range traceMap {
			allTraceNames[name] = true
		}
	}

	// Convert to slice and sort
	var traceNames []string
	for name := range allTraceNames {
		traceNames = append(traceNames, name)
	}
	sort.Strings(traceNames)

	// Summary table
	sb.WriteString("**Comparison Summary:**\n\n")
	sb.WriteString("| Trace Name |")
	for _, set := range traceSets {
		sb.WriteString(fmt.Sprintf(" %s |", getFileNameWithoutExt(set.Name)))
	}
	sb.WriteString(" Duration Diff |\n|------------")
	for range traceSets {
		sb.WriteString("|------------")
	}
	sb.WriteString("|------------|\n")

	// For each trace name, show if it exists in each set and calculate duration differences
	for _, name := range traceNames {
		sb.WriteString(fmt.Sprintf("| %s |", name))

		// Store durations for comparison
		var durations []time.Duration
		for _, traceMap := range traceMaps {
			if trace, exists := traceMap[name]; exists {
				sb.WriteString(" ✓ |")
				durations = append(durations, getTraceDuration(*trace))
			} else {
				sb.WriteString(" ✗ |")
				durations = append(durations, 0)
			}
		}

		// Calculate and show duration difference
		if len(durations) > 1 {
			firstDuration := durations[0]
			isSlowerThanAny := false
			var maxDiff time.Duration

			// Compare first duration with all others
			for i := 1; i < len(durations); i++ {
				if durations[i] > 0 { // Only compare with existing traces
					diff := durations[i] - firstDuration
					if diff < 0 {
						diff = -diff
					}
					if diff > maxDiff {
						maxDiff = diff
					}
					if firstDuration > durations[i] {
						isSlowerThanAny = true
					}
				}
			}

			if maxDiff > 0 {
				indicator := "🔴"
				if isSlowerThanAny {
					indicator = "🟢"
				}
				sb.WriteString(fmt.Sprintf(" %s %s |\n", indicator, formatDuration(maxDiff)))
			} else {
				sb.WriteString(" - |\n")
			}
		} else {
			sb.WriteString(" - |\n")
		}
	}
	sb.WriteString("\n")

	// Detailed comparison for matching traces
	sb.WriteString("**Detailed Comparison:**\n\n")
	for _, name := range traceNames {
		// Check if trace exists in all sets
		existsInAll := true
		for _, traceMap := range traceMaps {
			if _, exists := traceMap[name]; !exists {
				existsInAll = false
				break
			}
		}

		if existsInAll {
			sb.WriteString(fmt.Sprintf("<details>\n<summary>%s</summary>\n\n", name))

			// Show trace attributes
			sb.WriteString("**Trace Attributes:**\n\n")
			sb.WriteString("| Attribute |")
			for _, set := range traceSets {
				sb.WriteString(fmt.Sprintf(" %s |", getFileNameWithoutExt(set.Name)))
			}
			sb.WriteString("\n|-----------")
			for range traceSets {
				sb.WriteString("|-----------")
			}
			sb.WriteString("|\n")

			// Get all unique attribute keys
			allAttrKeys := make(map[string]bool)
			for _, traceMap := range traceMaps {
				trace := traceMap[name]
				for k := range trace.Attributes {
					allAttrKeys[k] = true
				}
				for k := range trace.ResourceAttrs {
					allAttrKeys[k] = true
				}
			}

			// Convert to slice and sort
			var attrKeys []string
			for k := range allAttrKeys {
				attrKeys = append(attrKeys, k)
			}
			sort.Strings(attrKeys)

			// Show attribute values for each set
			for _, key := range attrKeys {
				sb.WriteString(fmt.Sprintf("| %s |", key))
				for i, _ := range traceSets {
					trace := traceMaps[i][name]
					var value string
					if v, ok := trace.Attributes[key]; ok {
						value = v
					} else if v, ok := trace.ResourceAttrs[key]; ok {
						value = v
					}
					sb.WriteString(fmt.Sprintf(" %s |", value))
				}
				sb.WriteString("\n")
			}
			sb.WriteString("\n")

			// Compare spans
			sb.WriteString("**Span Comparison:**\n\n")
			sb.WriteString("| Span Name |")
			for _, set := range traceSets {
				sb.WriteString(fmt.Sprintf(" %s |", getFileNameWithoutExt(set.Name)))
			}
			sb.WriteString(" Duration Diff |\n|-----------")
			for range traceSets {
				sb.WriteString("|-----------")
			}
			sb.WriteString("|------------|\n")

			// Get all unique span names
			allSpanNames := make(map[string]bool)
			for _, traceMap := range traceMaps {
				trace := traceMap[name]
				for _, span := range trace.Spans {
					allSpanNames[span.Name] = true
				}
			}

			// Convert to slice and sort
			var spanNames []string
			for name := range allSpanNames {
				spanNames = append(spanNames, name)
			}
			sort.Strings(spanNames)

			// Show span durations for each set
			for _, spanName := range spanNames {
				sb.WriteString(fmt.Sprintf("| %s |", spanName))
				var spanDurations []time.Duration
				for i, _ := range traceSets {
					trace := traceMaps[i][name]
					var duration time.Duration
					found := false
					for _, span := range trace.Spans {
						if span.Name == spanName {
							duration = span.EndTime.Sub(span.StartTime)
							found = true
							break
						}
					}
					if found {
						sb.WriteString(fmt.Sprintf(" %s |", formatDuration(duration)))
						spanDurations = append(spanDurations, duration)
					} else {
						sb.WriteString(" ✗ |")
						spanDurations = append(spanDurations, 0)
					}
				}

				// Calculate and show duration difference for spans
				if len(spanDurations) > 1 {
					firstDuration := spanDurations[0]
					isSlowerThanAny := false
					var maxDiff time.Duration

					// Compare first duration with all others
					for i := 1; i < len(spanDurations); i++ {
						if spanDurations[i] > 0 { // Only compare with existing spans
							diff := spanDurations[i] - firstDuration
							if diff < 0 {
								diff = -diff
							}
							if diff > maxDiff {
								maxDiff = diff
							}
							if firstDuration > spanDurations[i] {
								isSlowerThanAny = true
							}
						}
					}

					if maxDiff > 0 {
						indicator := "🔴"
						if isSlowerThanAny {
							indicator = "🟢"
						}
						sb.WriteString(fmt.Sprintf(" %s %s |\n", indicator, formatDuration(maxDiff)))
					} else {
						sb.WriteString(" - |\n")
					}
				} else {
					sb.WriteString(" - |\n")
				}

				// Show span attributes
				sb.WriteString("| Attributes |")
				for i, _ := range traceSets {
					trace := traceMaps[i][name]
					var attrs []string
					for _, span := range trace.Spans {
						if span.Name == spanName {
							for k, v := range span.Attributes {
								attrs = append(attrs, fmt.Sprintf("%s: %s", k, v))
							}
							break
						}
					}
					sort.Strings(attrs)
					sb.WriteString(fmt.Sprintf(" %s |", strings.Join(attrs, "<br> ")))
				}
				sb.WriteString("\n")
			}

			sb.WriteString("\n</details>\n\n")
		}
	}

	return sb.String()
}
