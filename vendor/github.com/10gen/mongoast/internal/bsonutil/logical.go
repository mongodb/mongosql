package bsonutil

import "go.mongodb.org/mongo-driver/x/bsonx/bsoncore"

// And evaluates the logical AND operation on the two provided values.
func And(left, right bsoncore.Value) (bsoncore.Value, error) {
	if !CoerceToBoolean(left) || !CoerceToBoolean(right) {
		return False, nil
	}

	return True, nil
}

// Nor evaluates the logical NOR operation on the two provided values.
func Nor(left, right bsoncore.Value) (bsoncore.Value, error) {
	if CoerceToBoolean(left) || CoerceToBoolean(right) {
		return False, nil
	}

	return True, nil
}

// Not evaluates the logical NOT operation on the provided value.
func Not(v bsoncore.Value) (bsoncore.Value, error) {
	if CoerceToBoolean(v) {
		return False, nil
	}
	return True, nil
}

// Or evaluates the logical OR operation on the two provided values.
func Or(left, right bsoncore.Value) (bsoncore.Value, error) {
	if CoerceToBoolean(left) || CoerceToBoolean(right) {
		return True, nil
	}

	return False, nil
}
