package parser

const (
	// AvgAggregateName is a constant for the avg function.
	AvgAggregateName = "avg"
	// CountAggregateName is a constant for the count function.
	CountAggregateName = "count"
	// GroupConcatAggregateName is a constant for the group_concat function.
	GroupConcatAggregateName = "group_concat"
	// MaxAggregateName is a constant for the max function.
	MaxAggregateName = "max"
	// MinAggregateName is a constant for the min function.
	MinAggregateName = "min"
	// StdAggregateName is a constant for the std function.
	StdAggregateName = "std"
	// StdDevAggregateName is a constant for the stddev function.
	StdDevAggregateName = "stddev"
	// StdDevPopAggregateName is a constant for the stddev_pop function.
	StdDevPopAggregateName = "stddev_pop"
	// StdDevSampleAggregateName is a constant for the stddev_samp function.
	StdDevSampleAggregateName = "stddev_samp"
	// SumAggregateName is a constant for the sum function.
	SumAggregateName = "sum"
)

// AggregationFunctions is the set of aggregation function names.
var AggregationFunctions = map[string]struct{}{
	AvgAggregateName:          {},
	CountAggregateName:        {},
	GroupConcatAggregateName:  {},
	MaxAggregateName:          {},
	MinAggregateName:          {},
	StdAggregateName:          {},
	StdDevAggregateName:       {},
	StdDevPopAggregateName:    {},
	StdDevSampleAggregateName: {},
	SumAggregateName:          {},
}
