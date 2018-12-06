package mapping

import (
	"fmt"

	"github.com/10gen/sqlproxy/internal/config"
	"github.com/10gen/sqlproxy/schema/mongo"
)

// GetHeuristic gets the heuristic function for the given config.MappingMode.
func GetHeuristic(mode config.MappingMode) SchemataHeuristic {
	switch mode {
	case config.LatticeMappingMode:
		return PolymorphicTypeLatticeHeuristic
	case config.MajorityMappingMode:
		return PolymorphicMajorityCountHeuristic
	}
	panic(fmt.Sprintf("got schema mapping mode named '%s' which is invalid", mode))
}

// SchemataHeuristic is a function that chooses a dominant schema from a
// mongo.Schemata based on the mongo.Schemata's metadata
type SchemataHeuristic func(*mongo.Schemata) []*mongo.Schema

// CountHeuristic returns the Schema from the provided mongo.Schemata that has the
// highest count
var CountHeuristic = func(s *mongo.Schemata) []*mongo.Schema {
	var dominant *mongo.Schema
	var maxCount int

	for typ, sch := range s.Schemas {
		// a schema without a type should
		// never become dominant
		if typ == mongo.NoBSONType {
			continue
		}
		count := s.Counts[typ]

		var preferred bool
		if dominant == nil {
			preferred = true
		} else if count > maxCount {
			preferred = true
		} else if count == maxCount && typ.Less(dominant.BSONType) {
			preferred = true
		}

		if preferred {
			maxCount = count
			dominant = sch
		}
	}

	return []*mongo.Schema{dominant}
}

// bsonTypeAndSpecialType holds a BSONType and
// associated SpecialType.
type bsonTypeAndSpecialType struct {
	bsonType    mongo.BSONType
	specialType mongo.SpecialType
	count       int
}

type typeResolver = func([]bsonTypeAndSpecialType) bsonTypeAndSpecialType

// basePolymorphic Heuristic handles basic object/non-object conflicts
// and uses the passed scalarTypeResolver function to determine what scalar
// types to use.
func basePolymorphicHeuristic(scalarTypeResolver typeResolver, s *mongo.Schemata) []*mongo.Schema {
	scalarTypes := []bsonTypeAndSpecialType{{mongo.NoBSONType, mongo.NoSpecialType, 0}}
	var objectSchema *mongo.Schema
	var arraySchema *mongo.Schema
	hasScalarType := false
	for typ, sch := range s.Schemas {
		if typ == mongo.NoBSONType {
			continue
		}

		count := s.Counts[typ]
		switch typ {
		case mongo.Array:
			arraySchema = sch
		case mongo.Object:
			objectSchema = sch
		default:
			hasScalarType = true
			scalarTypes = append(scalarTypes, bsonTypeAndSpecialType{typ, sch.SpecialType, count})
		}
	}
	latticeType := scalarTypeResolver(scalarTypes)
	if arraySchema != nil {
		// Make sure to add the scalar latticeType as a schema
		// for the array Schema's items mongo.Schemata. Otherwise,
		// we will miss any scalar types that should bubble
		// up the array's item type (for instance, having a scalar
		// that samples with a string value when the array it
		// conflicts with has only double types, all the string
		// values would come out as 0 because the column would
		// be sampled as double). Do not bother to add mongo.NoBSONType,
		// even if we do, it should still work, but this is more
		// semantically correct.
		if _, ok := arraySchema.Items.Schemas[latticeType.bsonType]; !ok &&
			latticeType.bsonType != mongo.NoBSONType {
			arraySchema.Items.Schemas[latticeType.bsonType] = &mongo.Schema{
				BSONType:    latticeType.bsonType,
				SpecialType: latticeType.specialType,
			}
		}
		// If there is an objectSchema, we have an array/object conflict.
		if objectSchema != nil {
			return []*mongo.Schema{
				arraySchema,
				objectSchema,
			}
		}
		return []*mongo.Schema{arraySchema}
	}
	if objectSchema != nil {
		// If there is a scalar type also, we have a scalar/object conflict.
		if hasScalarType {
			return []*mongo.Schema{
				{
					BSONType:    latticeType.bsonType,
					SpecialType: latticeType.specialType,
				},
				objectSchema,
			}
		}
		return []*mongo.Schema{objectSchema}
	}
	return []*mongo.Schema{{
		BSONType:    latticeType.bsonType,
		SpecialType: latticeType.specialType,
	}}
}

// PolymorphicMajorityCountHeuristic handles polymorphic data. It merges
// scalar types based on the type lattice defined in
// https://docs.google.com/document/d/1FCsQ9ecDhQfamjvcgvfuaCNcW-RHAFNUdBTZpQWns_c/edit#
// treats array/scalar conflicts as being arrays, and x/object conflicts by returning
// two schemas, one for x and one for object.
var PolymorphicMajorityCountHeuristic = func(s *mongo.Schemata) []*mongo.Schema {
	return basePolymorphicHeuristic(getMajorityType, s)
}

// PolymorphicTypeLatticeHeuristic handles polymorphic data. It merges
// scalar types based on the type lattice defined in
// https://docs.google.com/document/d/1FCsQ9ecDhQfamjvcgvfuaCNcW-RHAFNUdBTZpQWns_c/edit#
// treats array/scalar conflicts as being arrays, and x/object conflicts by returning
// two schemas, one for x and one for object.
var PolymorphicTypeLatticeHeuristic = func(s *mongo.Schemata) []*mongo.Schema {
	return basePolymorphicHeuristic(getLeastUpperBound, s)
}

// getMajorityType selects the type based on whichever type was sampled most.
func getMajorityType(scalarTypes []bsonTypeAndSpecialType) bsonTypeAndSpecialType {
	if len(scalarTypes) == 1 {
		return scalarTypes[0]
	}
	maxTypes := []bsonTypeAndSpecialType{scalarTypes[0]}
	for _, ty := range scalarTypes[1:] {
		if ty.bsonType == mongo.NoBSONType {
			continue
		}
		if ty.count > maxTypes[0].count {
			maxTypes = []bsonTypeAndSpecialType{ty}
		} else if ty.count == maxTypes[0].count {
			maxTypes = append(maxTypes, ty)
		}
	}
	// If we have two or more types with the same count, resolve using
	// type lattice so that we will have schema stability if data
	// does not change between sampling.
	finalType := getLeastUpperBound(maxTypes)
	return finalType
}

// getLeastUpperBound resolves scalar type based on the least upper bound of all
// sampled scalar types according to the type lattice.
func getLeastUpperBound(scalarTypes []bsonTypeAndSpecialType) bsonTypeAndSpecialType {
	current := scalarTypes[0]
	if len(scalarTypes) == 1 {
		return current
	}
	for _, ty := range scalarTypes[1:] {
		current.bsonType = mongo.LeastUpperBound(current.bsonType, ty.bsonType)
		if ty.specialType != mongo.NoSpecialType {
			current.specialType = ty.specialType
		}
	}
	current.specialType = mongo.GetSpecialType(current.bsonType, current.specialType)
	return current
}
