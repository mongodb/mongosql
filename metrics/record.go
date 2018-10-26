package metrics

import (
	"encoding/json"
	"time"

	"github.com/10gen/sqlproxy/evaluator"
	"github.com/10gen/sqlproxy/parser"
	"github.com/satori/go.uuid"
)

const currentProtocol = 1

// Record represents a metrics record for a single query.
type Record interface {
	ToJSON() (string, error)
}

// NewRecord returns a new metrics.Record from the provided information.
func NewRecord(stmt parser.Statement, mongoVerson, biVersion string, stats *evaluator.PlanStats, latency int64) (Record, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	anonStmt := parser.AnonymizeStatement(stmt)

	rec := &record{
		ID:       id.String(),
		Protocol: currentProtocol,
		ExpireAt: getExpirationTime(),
		Query: queryRecord{
			SQL:  parser.String(anonStmt),
			Meta: newQueryMeta(parser.GetQueryStats(anonStmt)),
			Plan: newQueryPlan(stats),
			Execution: queryExecution{
				Success:   true,
				LatencyMS: latency,
			},
		},
		Variables: variableRecord{
			MongoVersion: mongoVerson,
			BIVersion:    biVersion,
		},
	}
	return rec, nil
}

type record struct {
	ID        string         `json:"_id"`
	Protocol  int            `json:"protocol"`
	ExpireAt  time.Time      `json:"expire_at"`
	Query     queryRecord    `json:"query"`
	Variables variableRecord `json:"variables"`
}

func (r *record) ToJSON() (string, error) {
	js, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(js), nil
}

func getExpirationTime() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year()+2, now.Month(), 1, 12, 0, 0, 0, time.UTC)
}

type variableRecord struct {
	MongoVersion string `json:"mongodb_version"`
	BIVersion    string `json:"mongosqld_version"`
}

type queryRecord struct {
	SQL       string         `json:"sql"`
	Meta      queryMeta      `json:"meta"`
	Plan      queryPlan      `json:"plan"`
	Execution queryExecution `json:"execution"`
}

type queryMeta struct {
	Functions  functionRecords  `json:"functions"`
	Joins      countKindRecords `json:"joins"`
	Unions     countKindRecords `json:"unions"`
	Subqueries countKindRecords `json:"subqueries"`
}

func newQueryMeta(stats *parser.QueryStats) queryMeta {
	return queryMeta{
		Functions:  functionRecords{funcs: stats.Functions},
		Joins:      countKindRecords{counts: stats.Joins},
		Unions:     countKindRecords{counts: stats.Unions},
		Subqueries: countKindRecords{counts: stats.Subqueries},
	}
}

type queryPlan struct {
	FullyPushedDown bool          `json:"fully_pushed_down"`
	Stages          []stageRecord `json:"stages"`
}

func newQueryPlan(stats *evaluator.PlanStats) queryPlan {
	stages := []stageRecord{}
	for _, rec := range stats.Explain {
		sr := newStageRecord(rec)
		stages = append(stages, sr)
	}

	return queryPlan{
		FullyPushedDown: stats.FullyPushedDown,
		Stages:          stages,
	}
}

type queryExecution struct {
	Success   bool  `json:"success"`
	LatencyMS int64 `json:"latency_ms"`
}

type stageRecord struct {
	ID        int    `json:"id"`
	StageType string `json:"stage_type"`
	Sources   []int  `json:"sources"`
}

func newStageRecord(rec *evaluator.ExplainRecord) stageRecord {
	return stageRecord{
		ID:        rec.ID,
		StageType: rec.StageType,
		Sources:   rec.Sources,
	}
}

type functionRecords struct {
	funcs map[string]int
}

func (r *functionRecords) MarshalJSON() ([]byte, error) {
	type funcRec = struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}
	out := []funcRec{}
	for f, c := range r.funcs {
		out = append(out, funcRec{f, c})
	}
	return json.Marshal(out)
}

type countKindRecords struct {
	counts map[string]int
}

func (r *countKindRecords) MarshalJSON() ([]byte, error) {
	type rec = struct {
		Kind  string `json:"kind"`
		Count int    `json:"count"`
	}
	out := []rec{}
	for k, c := range r.counts {
		out = append(out, rec{k, c})
	}
	return json.Marshal(out)
}
