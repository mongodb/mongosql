package mongo

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/10gen/mongo-go-driver/bson"
)

// BsonType is an enum representing all the bson types suported by Schema.
type BsonType string

// Possible values of BsonType
// for reference: https://docs.mongodb.com/manual/reference/operator/query/type/#available-types
const (
	Int        BsonType = "int"
	Long       BsonType = "long"
	Double     BsonType = "double"
	Decimal    BsonType = "decimal"
	Boolean    BsonType = "bool"
	String     BsonType = "string"
	Object     BsonType = "object"
	Array      BsonType = "array"
	Date       BsonType = "date"
	BinData    BsonType = "binData"
	ObjectId   BsonType = "objectId"
	NoBsonType BsonType = ""
)

// Less returns true if b should come before other when ordering BsonTypes.
// Less is equivalent to a lexicographical comparison of two BsonTypes' string
// representations.
func (b BsonType) Less(other BsonType) bool {
	return b < other
}

// SpecialType is an enum representing the possible values of the "specialType"
// field as descibed in the schema management design document. "specialType" is
// a sqlproxy-specific extension to JSON Schema that allows us to indicate that
// we want json-to-relation translation behavior different from the default for
// this schema's bsonType. For more information, see the design doc here:
// https://docs.google.com/document/d/12LWz00vJo_H-tHFv7IHa5L6X5a6Y-eNE4dyb8TXdk7U/edit
type SpecialType string

// Possible values of SpecialType
const (
	GeoPoint      SpecialType = "geoPoint"
	UUID3         SpecialType = "uuid3"
	UUID4         SpecialType = "uuid4"
	NoSpecialType SpecialType = ""
)

// Schema is a representation of a full or partial MongoDB schema. The Schema
// type can describe the schema for an object, array, literal, or a number of
// other complex types. Schema closely mirrors the structure of JSON Schema,
// with a few extensions for sqlproxy-specific needs.
type Schema struct {
	BsonType    BsonType             `json:"bsonType,omitempty" bson:"bsonType,omitempty"`
	Properties  map[string]*Schemata `json:"properties,omitempty" bson:"properties,omitempty"`
	Items       *Schemata            `json:"items,omitempty" bson:"items,omitempty"`
	SpecialType SpecialType          `json:"specialType,omitempty" bson:"specialType,omitempty"`
}

// NewCollectionSchema returns an empty Schema that can describe a collection.
// This function is intended to be the starting for most consumers of this
// package.
func NewCollectionSchema() *Schema {
	return &Schema{
		BsonType:   Object,
		Properties: make(map[string]*Schemata),
	}
}

// NewArraySchema returns a new Schema representing an array with the provided
// values as elements.
func NewArraySchema(values []interface{}) (*Schema, error) {
	// create empty Schemata
	schemata := NewSchemata(nil)

	// for each value, create a schema and add it to the schemata
	for _, v := range values {
		schema, err := NewSchemaFromValue(v)
		if err != nil {
			return nil, err
		}
		schemata.IncludeSchema(schema, 1)
	}

	// return the array schema
	return &Schema{
		BsonType: Array,
		Items:    schemata,
	}, nil
}

// NewEmptySchema returns a new schema that contains no information and places
// no constraints on the data it describes.
func NewEmptySchema() *Schema {
	return &Schema{}
}

// NewObjectSchema returns a Schema that describes the provided bson document.
func NewObjectSchema(doc bson.D) (*Schema, error) {
	props := make(map[string]*Schemata)

	// turn each doc element into a schemata
	for _, elem := range doc {
		propSchema, err := NewSchemaFromValue(elem.Value)
		if err != nil {
			return nil, err
		}
		props[elem.Name] = NewSchemata(propSchema)
	}

	return &Schema{
		BsonType:   Object,
		Properties: props,
	}, nil
}

// NewScalarSchema returns a Schema that describes a scalar value of the
// provided bson type.
func NewScalarSchema(t BsonType) (*Schema, error) {
	switch t {
	case Int, Long, Double, Decimal, Boolean, String, Date, BinData, ObjectId:
		return &Schema{
			BsonType: t,
		}, nil
	default:
		return nil, fmt.Errorf("Cannot create schema: invalid BsonType '%s'", t)
	}
}

// NewSchemaFromValue returns a new Schema whose type depends on the type of
// the provided value. If the provided value is unsupported, this function will
// return an error.
func NewSchemaFromValue(value interface{}) (*Schema, error) {
	switch typedV := value.(type) {
	case nil:
		// ignore
	case []interface{}:
		return NewArraySchema(typedV)
	case bson.D:
		return NewObjectSchema(typedV)
	case time.Time:
		return NewScalarSchema(Date)
	case bson.Binary:
		switch typedV.Kind {
		case 0x03:
			return NewUUIDSchema(UUID3)
		case 0x04:
			return NewUUIDSchema(UUID4)
		default:
			// ignore
		}
	case int, int32:
		return NewScalarSchema(Int)
	case int64:
		return NewScalarSchema(Long)
	case float32, float64:
		return NewScalarSchema(Double)
	case bson.Decimal128:
		return NewScalarSchema(Decimal)
	case string:
		return NewScalarSchema(String)
	case bool:
		return NewScalarSchema(Boolean)
	case bson.ObjectId:
		return NewScalarSchema(ObjectId)
	}

	return NewEmptySchema(), nil
}

// NewUUIDSchema returns a new Schema representing a UUID. The caller must pass
// a SpecialType that indicates the binary subtype for this UUID. If the
// provided SpecialType is not UUID3 or UUID4, an error will be returned.
func NewUUIDSchema(subtype SpecialType) (*Schema, error) {
	switch subtype {
	case UUID3, UUID4:
		return &Schema{
			BsonType:    BinData,
			SpecialType: subtype,
		}, nil
	default:
		return nil, fmt.Errorf("Cannot create a UUID Schema with binary subtype %s", subtype)
	}
}

// IncludeSample updates a Schema based on the provided document.
func (s *Schema) IncludeSample(doc bson.D) error {
	other, err := NewObjectSchema(doc)
	if err != nil {
		return err
	}
	return s.Merge(other)
}

// JsonSchema returns a JSON-Schema representation of a Schema in string form.
// The JSON-Schema representation returned by this function differs in a few
// important ways from the JSON-Schema standard. Those differences are discussed
// in this design document:
// https://docs.google.com/document/d/12LWz00vJo_H-tHFv7IHa5L6X5a6Y-eNE4dyb8TXdk7U/edit#
func (s *Schema) JsonSchema() (string, error) {
	b, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Merge combines the schema information from two Schemas into a single schema.
// Merge will return an error if the two schemas being merged are of different
// BsonTypes. Merging two scalar schemas will have no effect on the Schema.
// Merging two Array Schemas will combine their Item Schematas. Merging two
// Object schemas will yield an object whose set of keys is the union of the
// sets of keys of the two merged objects.
func (s *Schema) Merge(other *Schema) error {
	if s.BsonType != other.BsonType {
		return fmt.Errorf(
			"Cannot merge Schemas of differing types '%s' and '%s'",
			s.BsonType, other.BsonType)
	}

	switch s.BsonType {
	case Object:
		// merge each property in other into s
		for prop, otherSchemata := range other.Properties {

			// get the existing Schemata for this prop, or initialize if missing
			thisSchemata, ok := s.Properties[prop]
			if !ok {
				thisSchemata = NewSchemata(nil)
			}

			// merge the Schematas
			err := thisSchemata.Merge(otherSchemata)
			if err != nil {
				return err
			}

			// insert back into the map
			s.Properties[prop] = thisSchemata
		}

	case Array:
		return s.Items.Merge(other.Items)

	default:
		// if the schema type is unset or scalar, we have nothing to do

	}
	return nil
}

// Validate checks whether a Schema instance is a valid representation of a
// MongoDB schema. If the Schema's state is invalid, an error will be returned.
// Otherwise, this function returns nil.
func (s *Schema) Validate() error {
	switch s.BsonType {

	case Object:
		// Properties must be non-nil
		if s.Properties == nil {
			return fmt.Errorf("Properties must be non-nil for schema of BsonType %s", s.BsonType)
		}

		// Items must be nil
		if s.Items != nil {
			return fmt.Errorf("Items must be nil for schema of BsonType %s", s.BsonType)
		}

		// SpecialType must be GeoPoint or unset
		switch s.SpecialType {
		case NoSpecialType, GeoPoint:
			// these are allowed
		default:
			return fmt.Errorf("SpecialType %s invalid for schema of BsonType %s", s.SpecialType, s.BsonType)
		}

		// Schemata for each Property must be valid
		for prop, schemata := range s.Properties {
			err := schemata.Validate()
			if err != nil {
				return fmt.Errorf("Schemata for property '%s' failed validation: %s", prop, err.Error())
			}
		}

	case Array:
		// Properties must be nil
		if s.Properties != nil {
			return fmt.Errorf("Properties must be nil for schema of BsonType %s", s.BsonType)
		}

		// Items must be non-nil
		if s.Items == nil {
			return fmt.Errorf("Items must be non-nil for schema of BsonType %s", s.BsonType)
		}

		// SpecialType must be GeoPoint or unset
		switch s.SpecialType {
		case NoSpecialType, GeoPoint:
			// these are allowed
		default:
			return fmt.Errorf("SpecialType %s invalid for schema of BsonType %s", s.SpecialType, s.BsonType)
		}

		// Schemata for Items must be valid
		err := s.Items.Validate()
		if err != nil {
			return fmt.Errorf("Schemata for Items failed validation: %s", err.Error())
		}

	case BinData:
		// Properties must be nil
		if s.Properties != nil {
			return fmt.Errorf("Properties must be nil for schema of BsonType %s", s.BsonType)
		}

		// Items must be nil
		if s.Items != nil {
			return fmt.Errorf("Items must be nil for schema of BsonType %s", s.BsonType)
		}

		// SpecialType must be UUID3, UUID4, or unset
		switch s.SpecialType {
		case NoSpecialType, UUID3, UUID4:
			// these are allowed
		default:
			return fmt.Errorf("SpecialType %s invalid for schema of BsonType %s", s.SpecialType, s.BsonType)
		}

	case Int, Long, Double, Decimal, Boolean, String, Date, ObjectId, NoBsonType:
		// Properties must be nil
		if s.Properties != nil {
			return fmt.Errorf("Properties must be nil for schema of BsonType %s", s.BsonType)
		}

		// Items must be nil
		if s.Items != nil {
			return fmt.Errorf("Items must be nil for schema of BsonType %s", s.BsonType)
		}

		// SpecialType must be unset
		if s.SpecialType != NoSpecialType {
			return fmt.Errorf("SpecialType %s invalid for schema of BsonType %s", s.SpecialType, s.BsonType)
		}

	default:
		return fmt.Errorf("Invalid BsonType %s", s.BsonType)
	}

	return nil
}

// Schemata represents a superposition of multiple schemas. Schemata maintains
// one merged Schema for each unique BsonType added to it, along with some
// metadata that can be used to determine a "dominant" BsonType (and, by
// extension, Schema) for the Schemata.
type Schemata struct {
	Schemas   map[BsonType]*Schema
	Counts    map[BsonType]int
	Heuristic SchemataHeuristic
}

// SchemataHeuristic is a function that chooses a dominant schema from a
// Schemata based on the Schemata's metadata
type SchemataHeuristic func(*Schemata) *Schema

// CountHeuristic returns the Schema from the provided Schemata that has the
// highest count
var CountHeuristic = func(s *Schemata) *Schema {
	var dominant *Schema
	var maxCount int

	for typ, schema := range s.Schemas {
		count := s.Counts[typ]

		var preferred bool
		if dominant == nil {
			preferred = true
		} else if count > maxCount {
			preferred = true
		} else if count == maxCount && typ.Less(dominant.BsonType) {
			preferred = true
		}

		if preferred {
			maxCount = count
			dominant = schema
		}
	}

	return dominant
}

// NewSchemata returns a new Schemata containing only the provided Schema. If
// the provided Schema is nil, the returned Schemata will be empty.
func NewSchemata(s *Schema) *Schemata {
	schemas := make(map[BsonType]*Schema)
	counts := make(map[BsonType]int)

	if s != nil {
		schemas[s.BsonType] = s
		counts[s.BsonType] = 1
	}

	return &Schemata{
		Schemas:   schemas,
		Counts:    counts,
		Heuristic: CountHeuristic,
	}
}

// DominantSchema evaluates a Schemata according to its Heuristic, returning the
// schema that is returned by the Heuristic function.
func (s *Schemata) DominantSchema() *Schema {
	return s.Heuristic(s)
}

// GetBSON returns an object to be marshalled in place of the schemata when
// the schemata has to be marshalled. GetBSON will return the Schemata's
// dominant schema, as determined by Schemata.Heuristic.
func (s *Schemata) GetBSON() (interface{}, error) {
	return s.DominantSchema(), nil
}

// IncludeSchema will add the provided Schema in a Schemata. If the Schemata
// already contains a Schema of the same BsonType, the provided Schema will be
// merged with the existing Schema and the BsonType's count will be increased
// by the provided count. If a Schema of the provided type does not yet exist,
// the provided Schema will be used, with a starting count equal to the
// provided count.
func (s *Schemata) IncludeSchema(other *Schema, count int) error {

	// check if the schemata already has a schema of this type
	schemaType := other.BsonType
	schema, ok := s.Schemas[schemaType]

	if ok {
		// if so, increment the counter and merge the schemas
		s.Counts[schemaType] += count
		err := schema.Merge(other)
		if err != nil {
			return err
		}
	} else {
		// if not, add a new schema describing this sample to the schemata
		s.Counts[schemaType] = count
		s.Schemas[schemaType] = other
	}

	return nil
}

// MarshalJSON returns the JSON-Schema representation (in bytes) of this
// Schemata's dominant schema, as determined by Schemata.Heuristic.
func (s *Schemata) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.DominantSchema())
}

// Merge combines the Schema from two Schematas into a single Schemata. Merge
// is equivalent to calling IncludeSchema on each of other's Schemas.
func (s *Schemata) Merge(other *Schemata) error {
	for key, schema := range other.Schemas {
		count := other.Counts[key]
		err := s.IncludeSchema(schema, count)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetHeuristic allows a custom heuristic function to be used for choosing the
// dominant Schema from a Schemata.
func (s *Schemata) SetHeuristic(h SchemataHeuristic) {
	s.Heuristic = h
}

// Validate checks whether each Schema contained in this Schemata is valid.
// It will return the error from the first Schema that fails validation, or nil
// if all Schemas pass validation.
func (s *Schemata) Validate() error {
	for _, schema := range s.Schemas {
		err := schema.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}
