package evaluator

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/10gen/mongo-go-driver/bson"

	"github.com/10gen/sqlproxy/collation"
	"github.com/10gen/sqlproxy/internal/memory"
	"github.com/10gen/sqlproxy/internal/util"
	"github.com/10gen/sqlproxy/schema"
)

// UnionKind is an enum representing the different kinds of unions.
type UnionKind int

// These are the possible values for UnionKind.
const (
	UnionDistinct UnionKind = iota
	UnionAll
)

var (
	// hashSeed is defined by the FNV algorithm. It's a very large prime number.
	// see the wiki article for FNV hash: https://bit.ly/2Kcz4Ja
	hashSeed = util.Uint128{
		H: 0x6c62272e07bb0142,
		L: 0x62b821756295c58d,
	}
)

// UnionStage handles combining two result sets.
type UnionStage struct {
	left, right PlanStage
	kind        UnionKind
	is32        bool
}

// NewUnionStage creates a new UnionStage.
func NewUnionStage(kind UnionKind, left, right PlanStage) *UnionStage {
	return &UnionStage{
		kind:  kind,
		left:  left,
		right: right,
	}
}

// ensureFastPlanProjectInvariant makes sure to add _id: 0 to any "top level
// projects" that are leaves of the Union. This is necessary because the projects
// in the leaves of a union may not have the proper _id: 0 inserted. Not
// projecting away _id will break the fast iteration codepath, which requires that
// there be no columns outside of the columnInfo provided by the fastIter
// interface.
func ensureFastPlanProjectInvariant(fastPlan FastPlanStage) {
	if unionPlan, ok := fastPlan.(*UnionStage); ok {
		ensureFastPlanProjectInvariant(unionPlan.left.(FastPlanStage))
		ensureFastPlanProjectInvariant(unionPlan.right.(FastPlanStage))
		return
	}
	mongoSourcePlan, ok := fastPlan.(*MongoSourceStage)
	if !ok {
		panic(fmt.Sprintf("expected UnionStage or MongoSourceStage, but got :%T",
			fastPlan))
	}
	noIDDocElem := bson.DocElem{Name: mongoPrimaryKey, Value: 0}
	pipeline := mongoSourcePlan.pipeline
	if len(pipeline) == 0 {
		panic(fmt.Sprintf("expected pipeline with at least 1 stage,"+
			" got empty pipeline resulting from tables: %v",
			strings.Join(mongoSourcePlan.tableNames, ", ")))
	}
	lastStage := pipeline[len(pipeline)-1]
	if lastStage[0].Name != "$project" {
		// it is a coding error if the last stage of our pipeline
		// is not a project.
		panic(fmt.Sprintf("expected $project as last pipeline stage, got %v"+
			" resulting from tables: %v", lastStage[0].Name,
			strings.Join(mongoSourcePlan.tableNames, ", ")))
	}
	untypedProjectFields := lastStage[0].Value
	projectFields, ok := untypedProjectFields.(bson.D)
	if !ok {
		panic(fmt.Sprintf("expected bson.D in $project, got %T"+
			" resulting from tables %v",
			untypedProjectFields,
			strings.Join(mongoSourcePlan.tableNames, ", ")))
	}
	if _, ok := projectFields.Map()[mongoPrimaryKey]; !ok {
		projectFields = append(projectFields, noIDDocElem)
		lastStage[0].Value = projectFields
	}
}

// fastUnionPacket allows us to group together
// a bson.RawD with its accompanying columnInfo.
type fastUnionPacket struct {
	// columnInfo is a slice representing the field names,
	// in order, expected in the returned document.
	columnInfo []ColumnInfo
	// datum is the bson.RawD corresponding the a row from the
	// left or right side of a union.
	datum bson.RawD
}

// FastUnionAllIter returns BSON documents from the UNION ALL of two
// FastIters.
type FastUnionAllIter struct {
	// cancelIter allows for cancelling iteration.
	cancelIter context.CancelFunc
	// ctx is used to listen for any cancellation signals.
	ctx context.Context
	// columnInfo is a slice representing the field names,
	// in order, expected in the returned document.
	columnInfo []ColumnInfo
	// err holds any error that may occur during iteration.
	err error
	// errChan carries errors.
	errChan chan error
	// The left and right fast iter.
	left, right FastIter
	// leftChan and rightChan are channels returning bson.RawD results
	// for the left and right of the union, respectively.
	leftChan, rightChan chan fastUnionPacket
}

// FastUnionDistinctIter returns BSON documents from the UNION ALL of two
// FastIters.
type FastUnionDistinctIter struct {
	FastUnionAllIter
	// distinct set
	distinct map[util.Uint128]struct{}
	// stageMonitor is the memory monitor for the current stage.
	// FastUnionDistinct requires a map that introduces more
	// memory usage.
	stageMonitor *memory.Monitor
}

// FastUnionDistinctIter32 returns BSON documents from the UNION ALL of two
// FastIters when using server version < 3.4.0.
type FastUnionDistinctIter32 FastUnionDistinctIter

// UnionIter returns rows from the union of two source iterators.
type UnionIter struct {
	// cancelIter allows for cancelling iteration.
	cancelIter context.CancelFunc
	// columns holds the slice of Column structs for this iterator.
	columns []*Column
	// ctx is used to listen for any cancellation signals.
	ctx *ExecutionCtx
	// err holds any error that may occur during iteration.
	err error
	// errChan carries errors.
	errChan chan error
	// The left and right iter.
	left, right Iter
	// onChan returns the unified row results.
	onChan chan Row
	// stageMonitor is the memory monitor for the current stage.
	stageMonitor *memory.Monitor
}

// FastOpen opens a FastIter for the UnionStage.
func (union *UnionStage) FastOpen(ctx *ExecutionCtx) (FastIter, error) {
	fastPlanLeft, ok := union.left.(FastPlanStage)
	if !ok {
		panic(fmt.Sprintf("left child of UnionStage must be FastPlanStage, "+
			"got %T", union.left))
	}
	fastPlanRight, ok := union.right.(FastPlanStage)
	if !ok {
		panic(fmt.Sprintf("right child of UnionStage must be FastPlanStage, "+
			"got %T", union.right))
	}
	// Fix projects since they are not correctly optimized as "top level".
	// This means making sure _id: 0 is set.
	ensureFastPlanProjectInvariant(fastPlanLeft)
	ensureFastPlanProjectInvariant(fastPlanRight)

	cancelCtx, cancel := context.WithCancel(ctx.Context())

	iter := &FastUnionAllIter{
		ctx:        ctx.Context(),
		err:        nil,
		errChan:    make(chan error, 2),
		cancelIter: cancel,
	}

	initErrChan := make(chan error, 2)
	initDoneChan := make(chan struct{}, 2)

	handleError := func(errChan chan error) func(err interface{}) {
		return func(err interface{}) {
			errChan <- fmt.Errorf("%v", err)
		}
	}

	util.PanicSafeGo(func() {
		fastIterLeft, err := fastPlanLeft.FastOpen(ctx)
		if err != nil {
			initErrChan <- err
			return
		}
		iter.left = fastIterLeft
		initDoneChan <- struct{}{}
	}, handleError(initErrChan))

	util.PanicSafeGo(func() {
		fastIterRight, err := fastPlanRight.FastOpen(ctx)
		if err != nil {
			initErrChan <- err
			return
		}
		iter.right = fastIterRight
		initDoneChan <- struct{}{}
	}, handleError(initErrChan))

	// Wait for initialization.
	for doneCount := 0; doneCount < 2; {
		select {
		case err := <-initErrChan:
			return nil, err
		case <-initDoneChan:
			doneCount++
		}
	}

	iter.columnInfo = iter.right.GetColumnInfo()

	iter.leftChan = make(chan fastUnionPacket)
	iter.rightChan = make(chan fastUnionPacket)

	iterateSide := func(it FastIter, channel chan fastUnionPacket) func() {
		return func() {
			b := &bson.RawD{}
		Loop:
			for it.Next(b) {
				channel <- fastUnionPacket{
					columnInfo: it.GetColumnInfo(),
					datum:      *b,
				}
				select {
				case <-cancelCtx.Done():
					break Loop
				default:
				}
				b = &bson.RawD{}
			}
			close(channel)
		}
	}

	util.PanicSafeGo(iterateSide(iter.left, iter.leftChan), handleError(iter.errChan))
	util.PanicSafeGo(iterateSide(iter.right, iter.rightChan), handleError(iter.errChan))

	if union.kind == UnionDistinct {
		stageMonitor, err := newStageMemoryMonitor(ctx, "FastUnionDistinctStage")
		if err != nil {
			return nil, err
		}
		if union.is32 {
			return &FastUnionDistinctIter32{
				FastUnionAllIter: *iter,
				distinct:         make(map[util.Uint128]struct{}),
				stageMonitor:     stageMonitor,
			}, nil
		}
		return &FastUnionDistinctIter{
			FastUnionAllIter: *iter,
			distinct:         make(map[util.Uint128]struct{}),
			stageMonitor:     stageMonitor,
		}, nil
	}
	return iter, nil
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (iter *FastUnionAllIter) Next(doc *bson.RawD) bool {
	getNextPacket := func(packet fastUnionPacket,
		ok bool,
		otherChan chan fastUnionPacket) bool {
		if ok {
			iter.columnInfo = packet.columnInfo
			*doc = packet.datum
			return true
		}
		packet, ok = <-otherChan
		if !ok {
			return false
		}
		iter.columnInfo = packet.columnInfo
		*doc = packet.datum
		return true
	}
	select {
	case err := <-iter.errChan:
		iter.err = err
		return false
	case p, ok := <-iter.leftChan:
		return getNextPacket(p, ok, iter.rightChan)
	case p, ok := <-iter.rightChan:
		return getNextPacket(p, ok, iter.leftChan)
	}
}

func addValueToHash(hash util.Uint128, value bson.Raw) util.Uint128 {
	hash.AddByteToHash(value.Kind)
	hash.AddByteSliceToHash(value.Data)
	return hash
}

// computeHash computes and FNV-1a (plus prime) hash for a bson.RawD.
func (iter *FastUnionDistinctIter) computeHash(datum *bson.RawD) util.Uint128 {
	columnInfo := iter.columnInfo
	values := *datum
	lenColumnFields, lenValues := len(columnInfo), len(values)
	// if lengths are equal, we have no missing values.
	if lenColumnFields == lenValues {
		hash := hashSeed
		for _, val := range values {
			// We need to add the kind to the hash as well, as NULLs
			// are stored with 0 bytes.
			hash = addValueToHash(hash, val.Value)
		}
		return hash
	}
	// If we have missing fields, we need to check key names. Until
	// we have determined all the missing fields.
	// This can be simplified by removing missing fields from returns.
	numMissingValues := lenColumnFields - lenValues
	hash := hashSeed
	i := 0
	for _, info := range columnInfo {
		fieldName := info.Field
		if numMissingValues > 0 && i < lenValues {
			if fieldName == values[i].Name {
				value := values[i].Value
				// If this is the correct fieldName, output the value.
				hash = addValueToHash(hash, value)
				// increment i so that we consider the next value.
				i++
			} else {
				// If the fieldName is wrong, this field must be missing, output
				// a NULL, decrement numMissingValues (because we found one), but do NOT
				// touch i because we want the same position in the values next
				// iteration.
				hash.AddByteToHash(byte(schema.BSONNull))
				numMissingValues--
			}
		} else if i < len(values) {
			// We have found all the missing values, default to the faster mode.
			value := values[i].Value
			i++
			hash = addValueToHash(hash, value)
		} else {
			// i >= len(values), break to where we add NULLS, if necessary.
			break
		}
	}
	// We ran out of values, all values after this point must be missing.
	for ; numMissingValues != 0; numMissingValues-- {
		hash.AddByteToHash(byte(schema.BSONNull))
	}
	return hash
}

// computeHash computes and FNV-1a (plus prime) hash for a bson.RawD on server versions < 3.4.0.
func (iter *FastUnionDistinctIter32) computeHash(datum *bson.RawD) util.Uint128 {
	values := *datum
	columnInfo := iter.GetColumnInfo()
	lenColumnInfo := len(columnInfo)
	// We will use one nullField value to represent all NULLs that will result
	// from missing fields.
	nullField := bson.Raw{Kind: byte(schema.BSONNull), Data: []byte{}}
	fieldMap := make(map[string]bson.Raw, lenColumnInfo)
	// Set the value for all columns to null so we can avoid
	// a branch in the loop below.
	for _, info := range columnInfo {
		fieldMap[info.Field] = nullField
	}
	// We can't rely on field ordering in 3.2.
	for i := range values {
		fieldMap[values[i].Name] = values[i].Value
	}
	hash := hashSeed
	for _, info := range columnInfo {
		value := fieldMap[info.Field]
		hash = addValueToHash(hash, value)
	}
	return hash
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used. This
// is the only method that differs for Distinct Union, in that we check for
// duplicates.
func (iter *FastUnionDistinctIter) Next(doc *bson.RawD) bool {
	for {
		if !iter.FastUnionAllIter.Next(doc) {
			return false
		}
		hash := iter.computeHash(doc)
		if _, ok := iter.distinct[hash]; !ok {
			iter.distinct[hash] = struct{}{}
			// each util.Uint128 in the map is 16 bytes.
			err := iter.stageMonitor.Acquire(16)
			if err != nil {
				iter.errChan <- err
				return false
			}
			return true
		}
	}
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used. This
// is the only method that differs for Distinct Union, in that we check for
// duplicates. This version is used for MongoDB Version 3.2. Unfortunately, we need
// to have duplicated code for the correct computeHash to be called.
func (iter *FastUnionDistinctIter32) Next(doc *bson.RawD) bool {
	for {
		if !iter.FastUnionAllIter.Next(doc) {
			return false
		}
		hash := iter.computeHash(doc)
		if _, ok := iter.distinct[hash]; !ok {
			iter.distinct[hash] = struct{}{}
			// each util.Uint128 in the map is 16 bytes.
			err := iter.stageMonitor.Acquire(16)
			if err != nil {
				iter.errChan <- err
				return false
			}
			return true
		}
	}
}

// GetColumnInfo returns the slice of ColumnInfo necessary for streaming the results.
func (iter *FastUnionAllIter) GetColumnInfo() []ColumnInfo {
	return iter.columnInfo
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (iter *FastUnionAllIter) Err() error {
	if err := iter.left.Err(); err != nil {
		return err
	}

	if err := iter.right.Err(); err != nil {
		return err
	}

	return iter.err
}

// Close closes the iterator, returning any error encountered while doing so.
func (iter *FastUnionAllIter) Close() error {
	iter.cancelIter()

	err := iter.left.Close()
	if err != nil {
		return err
	}

	return iter.right.Close()
}

// Close closes the iterator, and releases the memory monitor memory.
func (iter *FastUnionDistinctIter) Close() error {
	// each item in the map is a util.Uint128, which means 16 bytes.
	err := iter.stageMonitor.Release(16 * uint64(len(iter.distinct)))
	if err != nil {
		return err
	}
	// release the distinct map asap.
	iter.distinct = nil
	return iter.FastUnionAllIter.Close()
}

// Open returns an iterator that returns results from executing this plan stage
// with the given ExecutionContext.
func (union *UnionStage) Open(ctx *ExecutionCtx) (Iter, error) {
	stageMonitor, err := newStageMemoryMonitor(ctx, "UnionStage")
	if err != nil {
		return nil, err
	}

	cancelCtx, cancel := context.WithCancel(ctx.Context())

	iter := &UnionIter{
		ctx:          ctx,
		stageMonitor: stageMonitor,
		columns:      union.Columns(),
		errChan:      make(chan error, 1),
		cancelIter:   cancel,
	}

	leftRows := make(chan *Row)
	rightRows := make(chan *Row)

	util.PanicSafeGo(func() {
		iterator, err := union.left.Open(ctx)
		if err != nil {
			iter.errChan <- err
			return
		}
		iter.left = iterator
		iter.fetchRows(cancelCtx, iterator, leftRows, iter.errChan)
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("%v", err)
	})

	util.PanicSafeGo(func() {
		iterator, err := union.right.Open(ctx)
		if err != nil {
			iter.errChan <- err
			return
		}
		iter.right = iterator
		iter.fetchRows(cancelCtx, iterator, rightRows, iter.errChan)
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("%v", err)
	})

	iter.onChan = iter.unify(cancelCtx, leftRows, rightRows)

	return iter, nil
}

func (iter *UnionIter) fetchRows(ctx context.Context, it Iter, ch chan *Row, errChan chan error) {
	r := &Row{}

	syncChan := make(chan *Row)
	fetchErrChan := make(chan error, 1)

	util.PanicSafeGo(func() {
		for it.Next(r) {

			inSize := r.Data.Size()

			err := iter.stageMonitor.Include(inSize)
			if err != nil {
				errChan <- err
				break
			}

			// Need to match row info with parent
			for i, col := range iter.columns {
				r.Data[i].Name = col.Name
				r.Data[i].Data, _ = NewSQLValue(r.Data[i].Data, col.SQLType, schema.SQLNone)
			}

			err = iter.stageMonitor.Release(inSize)
			if err != nil {
				errChan <- err
				break
			}

			err = iter.stageMonitor.Acquire(r.Data.Size())
			if err != nil {
				errChan <- err
				break
			}

			select {
			case syncChan <- r:
				r = &Row{}
			case <-ctx.Done():
			}
		}

		if err := it.Err(); err != nil {
			errChan <- err
		}

		// This err was previously ignored.
		if err := it.Close(); err != nil {
			panic(err)
		}

		close(syncChan)
	}, func(err interface{}) {
		fetchErrChan <- fmt.Errorf("union fetch error: %v", err)
	})

	for {
		select {
		case row, ok := <-syncChan:
			if !ok {
				close(ch)
				return
			}

			ch <- row
		case <-ctx.Done():
			errChan <- ctx.Err()
			return
		case err := <-fetchErrChan:
			errChan <- err
			return
		}
	}
}

func mergeColumnsByType(lcols, rcols []*Column) []*Column {
	outCols := make([]*Column, len(lcols))

	sorter := &schema.SQLTypesSorter{}
	for i, lcol := range lcols {
		rcol := rcols[i]
		sorter.Types = []schema.SQLType{lcol.SQLType, rcol.SQLType}
		sort.Sort(sorter)

		outCol := lcol.clone()
		outCol.SQLType = sorter.Types[1] // Use "gte" type
		outCols[i] = outCol
	}

	return lcols
}

// Columns returns the ordered set of columns that are contained in results from this plan.
func (union *UnionStage) Columns() []*Column {
	return mergeColumnsByType(union.left.Columns(), union.right.Columns())
}

// Collation returns the collation to use for comparisons.
func (union *UnionStage) Collation() *collation.Collation {
	return union.left.Collation()
}

// Next populates the provided Row with this iterator's next available row.
// If the iterator has been exhausted or has encountered an error, Next will
// return false, and the value of the provided Row should not be used.
func (iter *UnionIter) Next(row *Row) bool {
	select {
	case err := <-iter.errChan:
		iter.err = err
		return false
	case data, ok := <-iter.onChan:
		row.Data = data.Data
		if !ok {
			return false
		}

		iter.err = iter.stageMonitor.Release(row.Data.Size())
		if iter.err != nil {
			return false
		}

		// past this stage, all columns must
		// present the same table name.
		for i := 0; i < len(row.Data); i++ {
			row.Data[i].Table = iter.columns[i].Table
		}

		iter.err = iter.stageMonitor.Acquire(row.Data.Size())
		if iter.err != nil {
			return false
		}
		iter.err = iter.stageMonitor.Exclude(row.Data.Size())
		if iter.err != nil {
			return false
		}
	}
	return true
}

// Close closes the iterator, returning any error encountered while doing so.
func (iter *UnionIter) Close() error {
	iter.cancelIter()

	err := iter.left.Close()
	if err != nil {
		return err
	}

	err = iter.right.Close()
	if err != nil {
		return err
	}

	_, err = iter.stageMonitor.Clear()
	return err
}

// Err returns any error that has been encountered while iterating. If no error
// was encountered, Err returns nil.
func (iter *UnionIter) Err() error {

	if err := iter.left.Err(); err != nil {
		return err
	}

	if err := iter.right.Err(); err != nil {
		return err
	}

	return iter.err
}

func (iter *UnionIter) unify(ctx context.Context, lChan, rChan chan *Row) chan Row {

	ch := make(chan Row)
	closeChan := make(chan struct{})

	// cleanup
	util.PanicSafeGo(func() {
		<-closeChan
		<-closeChan
		close(closeChan)
		close(ch)
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("%v", err)
	})

	// retrieve rows from left and right stages in parallel
	util.PanicSafeGo(func() {
	chanLoop:
		for l := range lChan {
			select {
			case ch <- *l:
			case <-ctx.Done():
				break chanLoop
			}
		}
		closeChan <- struct{}{}
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("left unify error: %v", err)
	})

	util.PanicSafeGo(func() {
	chanLoop:
		for r := range rChan {
			select {
			case ch <- *r:
			case <-ctx.Done():
				break chanLoop
			}
		}
		closeChan <- struct{}{}
	}, func(err interface{}) {
		iter.errChan <- fmt.Errorf("right unify error: %v", err)
	})

	return ch
}
