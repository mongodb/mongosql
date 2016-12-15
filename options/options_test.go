package options_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/10gen/sqlproxy/options"
	. "github.com/smartystreets/goconvey/convey"
)

func TestEnsureOptsNotNil(t *testing.T) {
	Convey("Subject: EnsureOptsNotNil", t, func() {
		myOpts, err := options.NewSqldOptions()
		So(err, ShouldBeNil)
		defaultMaskFlags := []string{"addr", "filePermissions", "unixSocketPrefix", "mongo-uri"}
		defaultMaskValues := []string{"127.0.0.1:3307", "0700", "/tmp", "mongodb://localhost:27017"}

		options.EnsureOptsNotNil(&myOpts)
		for i, opts := 0, reflect.ValueOf(&myOpts).Elem(); i < opts.NumField(); i++ {
			// Skip any non-exported fields
			if !opts.Field(i).Elem().CanSet() {
				continue
			}
			So(opts.Field(i).Interface(), ShouldNotBeNil)
			for j, subOpts := 0, opts.Field(i).Elem(); j < subOpts.NumField(); j++ {
				if subOpts.Field(j).Kind() == reflect.Ptr {
					So(subOpts.Field(j).Interface(), ShouldNotBeNil)
					// Ensure that the values of elements in defaultMaskFlags are equal to their
					// corresponding defaultMaskValues
					for index, _ := range defaultMaskFlags {
						if subOpts.Type().Field(j).Tag.Get("long") == defaultMaskFlags[index] {
							So(subOpts.Field(j).Elem().Interface(), ShouldEqual, defaultMaskValues[index])
						}
					}
				}
			}
		}
	})
}

func populateOpts(opts *options.SqldOptions, args map[string]string) {
	for field, arg := range args {
		fieldVal := reflect.ValueOf(opts).Elem().FieldByName(field)
		if !fieldVal.IsValid() {
			panic(fmt.Sprintf("Expected %v to be in SqldOptions but did not find it", field))
		}
		if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
			fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
			switch fieldVal.Elem().Kind() {
			case reflect.Bool:
				argBool, _ := strconv.ParseBool(arg)
				fieldVal.Elem().Set(reflect.ValueOf(argBool))
			case reflect.String:
				fieldVal.Elem().Set(reflect.ValueOf(arg))
			case reflect.Int:
				argInt, _ := strconv.Atoi(arg)
				fieldVal.Elem().Set(reflect.ValueOf(argInt))
			default:
				panic(fmt.Sprintf("%v points to unexpected type", field))
			}
		}
	}
}

func TestSqldOptionsString(t *testing.T) {
	Convey("Subject: SqldOptions string()", t, func() {
		Convey("Should redact both MongoPEMKeyPassword and SSLPEMKeyFilePassword", func() {
			myOpts, err := options.NewSqldOptions()
			So(err, ShouldBeNil)
			myArgs := map[string]string{
				"SSLPEMKeyFilePassword":   "password123",
				"MongoPEMKeyFilePassword": "password123",
			}
			populateOpts(&myOpts, myArgs)
			So(myOpts.String(), ShouldEqual, "--sslPEMKeyPassword <password> --mongo-sslPEMKeyPassword <password>")
		})
		Convey("Should properly stringify both string and non-string argument values", func() {
			myOpts, err := options.NewSqldOptions()
			So(err, ShouldBeNil)
			myArgs := map[string]string{
				"SSLPEMKeyFilePassword":   "password123",
				"MongoPEMKeyFilePassword": "password123",
				"UnixSocketPrefix":        "/dir",
				"NoUnixSocket":            "true",
				"Addr":                    "127.0.0.1:3308",
				"VLevel":                  "5",
			}
			populateOpts(&myOpts, myArgs)
			So(myOpts.String(), ShouldEqual, "--addr 127.0.0.1:3308 --sslPEMKeyPassword <password> --verbose 5 --mongo-sslPEMKeyPassword <password> --noUnixSocket true --unixSocketPrefix /dir")
		})
	})
}
