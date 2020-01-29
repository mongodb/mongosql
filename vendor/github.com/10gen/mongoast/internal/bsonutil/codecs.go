package bsonutil

import (
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

var Registry = func() *bsoncodec.Registry {
	rb := bson.NewRegistryBuilder()
	register(rb)
	return rb.Build()
}()

func register(rb *bsoncodec.RegistryBuilder) {
	rb.RegisterDecoder(tCoreValue, bsoncodec.ValueDecoderFunc(bsoncoreValueDecodeValue))
	rb.RegisterEncoder(tCoreValue, bsoncodec.ValueEncoderFunc(bsoncoreValueEncodeValue))
}

var tCoreValue = reflect.TypeOf(bsoncore.Value{})

func bsoncoreValueDecodeValue(dc bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != tCoreValue {
		return bsoncodec.ValueDecoderError{Name: "bsoncoreValueDecodeValue", Types: []reflect.Type{tCoreValue}, Received: val}
	}

	t, value, err := bsonrw.Copier{}.CopyValueToBytes(vr)
	if err != nil {
		return err
	}

	val.Set(reflect.ValueOf(bsoncore.Value{Type: t, Data: value}))
	return nil
}

func bsoncoreValueEncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tCoreValue {
		return bsoncodec.ValueEncoderError{Name: "bsoncoreValueEncodeValue", Types: []reflect.Type{tCoreValue}, Received: val}
	}

	v := val.Interface().(bsoncore.Value)

	return bsonrw.Copier{}.CopyValueFromBytes(vw, v.Type, v.Data)
}
