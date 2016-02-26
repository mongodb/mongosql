package evaluator

import "fmt"

func OptimizeOperator(ctx *ExecutionCtx, o Operator) (Operator, error) {
	v := &optimizer{ctx}
	return v.Visit(o)
}

type optimizer struct {
	ctx *ExecutionCtx
}

func (v *optimizer) Visit(o Operator) (Operator, error) {

	o, err := walkOperatorTree(v, o)
	if err != nil {
		return nil, err
	}

	switch typedO := o.(type) {
	case *Filter:
		o, err = v.visitFilter(typedO)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize filter: %v", err)
		}
	case *GroupBy:
		o, err = v.visitGroupBy(typedO)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize group by: %v", err)
		}
	case *Join:
		o, err = v.visitJoin(typedO)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize join: %v", err)
		}
	case *Limit:
		o, err = v.visitLimit(typedO)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize limit: %v", err)
		}
	case *OrderBy:
		o, err = v.visitOrderBy(typedO)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize order by: %v", err)
		}
	case *Project:
		o, err = v.visitProject(typedO)
		if err != nil {
			return nil, fmt.Errorf("unable to optimize project: %v", err)
		}
	}

	return o, nil
}

func canPushDown(op Operator) (*MongoSource, bool) {

	ms, ok := op.(*MongoSource)
	if !ok {
		return nil, false
	}

	return ms, true
}
