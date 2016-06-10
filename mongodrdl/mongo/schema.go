package mongo

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"time"
)

const (
	ArraySchemaTypeName    = "array"
	DocumentSchemaTypeName = "document"
)

// +++++++++++++++++++++
type Type interface {
	Name() string
	Count() int
	Copy() Type
	Combine(Type) // increments the count
}

func (*Array) Name() string    { return ArraySchemaTypeName }
func (*Document) Name() string { return DocumentSchemaTypeName }
func (s *Scalar) Name() string { return s.name }
func (a *Array) Count() int    { return a.count }
func (d *Document) Count() int { return d.count }
func (s *Scalar) Count() int   { return s.count }
func (a *Array) Combine(with Type) {
	a2 := with.(*Array)
	a.count += a2.count
	a2.TypeContainer.copyTo(&a.TypeContainer)
}
func (d *Document) Combine(with Type) {
	d2 := with.(*Document)
	for key, f2 := range d2.Fields {
		f1, exists := d.Fields[key]
		if exists {
			f1.combine(f2)
		} else {
			d.Fields[key] = f2.copy()
		}
	}
}
func (s *Scalar) Combine(with Type) {
	s2 := with.(*Scalar)
	s.count += s2.count
}
func (a *Array) Copy() Type {
	newA := NewArray()
	a.TypeContainer.copyTo(&newA.TypeContainer)
	newA.count = a.count
	return newA
}
func (d *Document) Copy() Type {
	newD := NewDocument()
	for key, f := range d.Fields {
		newD.Fields[key] = f.copy()
	}
	newD.count = d.count
	return newD
}
func (s *Scalar) Copy() Type {
	newS := NewScalar(s.name)
	newS.count = s.count
	return newS
}

// +++++++++++++++++++++
type TypeContainer struct {
	Types []Type
}

func (c *TypeContainer) copyTo(dest *TypeContainer) {
	for _, t := range c.Types {
		destT, exists := dest.getByName(t.Name())
		if exists {
			destT.Combine(t)
		} else {
			dest.Types = append(dest.Types, t.Copy())
		}
	}
}

func (c *TypeContainer) getByName(name string) (Type, bool) {
	for _, t := range c.Types {
		if t.Name() == name {
			return t, true
		}
	}

	return nil, false
}

func (c *TypeContainer) getOrAddArray() *Array {
	t, exists := c.getByName(ArraySchemaTypeName)
	if !exists {
		t = NewArray()
		c.Types = append(c.Types, t)
	}

	return t.(*Array)
}

func (c *TypeContainer) getOrAddDocument() *Document {
	t, exists := c.getByName(DocumentSchemaTypeName)
	if !exists {
		t = NewDocument()
		c.Types = append(c.Types, t)
	}

	return t.(*Document)
}

func (c *TypeContainer) getOrAddScalar(name string) *Scalar {
	t, exists := c.getByName(name)
	if !exists {
		t = NewScalar(name)
		c.Types = append(c.Types, t)
	}

	return t.(*Scalar)
}

func (c *TypeContainer) includeSample(value interface{}) error {
	switch v := value.(type) {
	case []interface{}:
		array := c.getOrAddArray()
		sample, success := value.([]interface{})
		if !success {
			return fmt.Errorf("Could not cast value: %v", value)
		}
		err := array.includeSample(sample)
		if err != nil {
			return err
		}
	case bson.D:
		doc := c.getOrAddDocument()
		sample, success := value.(bson.D)
		if !success {
			return fmt.Errorf("Could not cast value to *bson.D: %v", value)
		}
		err := doc.includeSample(sample)
		if err != nil {
			return err
		}
	case nil:
		// ignore nil values since they have no type information to process
	case time.Time:
		s := c.getOrAddScalar("date")
		s.includeSample()
	default:
		s := c.getOrAddScalar(fmt.Sprintf("%T", v))
		s.includeSample()
	}

	return nil
}

// +++++++++++++++++++++
type Array struct {
	TypeContainer
	count int
}

func NewArray() *Array {
	return &Array{}
}

func (a *Array) SetCount(count int) {
	a.count = count
}

func (a *Array) includeSample(values []interface{}) error {
	a.count++
	for _, v := range values {
		err := a.TypeContainer.includeSample(v)
		if err != nil {
			return err
		}
	}
	return nil
}

// +++++++++++++++++++++
type Collection struct {
	Document
	Name string
}

func NewCollection(name string) *Collection {
	c := &Collection{
		Name: name,
	}

	c.Document.Fields = make(map[string]*Field)
	return c
}

func (c *Collection) IncludeSample(doc bson.D) error {
	return c.Document.includeSample(doc)
}

// +++++++++++++++++++++
type Document struct {
	Fields map[string]*Field
	count  int
}

func NewDocument() *Document {
	return &Document{
		Fields: make(map[string]*Field),
	}
}

func CombineDocuments(d1 *Document, d2 *Document) *Document {
	// make a copy of the first guy...
	newD := d1.Copy().(*Document)
	newD.Combine(d2)
	return newD
}

func (d *Document) includeSample(doc bson.D) error {
	for _, elem := range doc {
		f := d.getOrAddField(elem.Name)

		err := f.includeSample(elem.Value)
		if err != nil {
			return err
		}
	}

	d.count++
	return nil
}

func (d *Document) getOrAddField(name string) *Field {
	field, exists := d.Fields[name]
	if !exists {
		field = NewField(name)
		d.Fields[name] = field
	}

	return field
}

// +++++++++++++++++++++
type Scalar struct {
	name  string
	count int
}

func NewScalar(name string) *Scalar {
	return &Scalar{
		name: name,
	}
}

func (s *Scalar) includeSample() {
	s.count++
}

// +++++++++++++++++++++
type Field struct {
	TypeContainer
	Name  string
	Count int
}

func NewField(name string) *Field {
	field := &Field{
		Name: name,
	}
	return field
}

func (f *Field) combine(f2 *Field) {
	f2.TypeContainer.copyTo(&f.TypeContainer)
	f.Count += f2.Count
}

func (f *Field) copy() *Field {
	newF := NewField(f.Name)
	f.TypeContainer.copyTo(&newF.TypeContainer)
	newF.Count = f.Count
	return newF
}

func (f *Field) includeSample(value interface{}) error {
	f.Count++
	return f.TypeContainer.includeSample(value)
}
