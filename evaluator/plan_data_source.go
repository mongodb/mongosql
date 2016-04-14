package evaluator

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type FindResults interface {
	Next(result *bson.D) bool
	Err() error
	Close() error
}

type FindQuery interface {
	Iter() FindResults
}

type MgoFindResults struct {
	iter *mgo.Iter
}

func (mfr MgoFindResults) Next(result *bson.D) bool {
	return mfr.iter.Next(result)
}

func (mfr MgoFindResults) Err() error {
	return mfr.iter.Err()
}

func (mfr MgoFindResults) Close() error {
	return mfr.iter.Close()
}

type EmptyFindResults struct {
}

func (_ EmptyFindResults) Next(result *bson.D) bool {
	return false
}

func (_ EmptyFindResults) Err() error {
	return nil
}

func (_ EmptyFindResults) Close() error {
	return nil
}
