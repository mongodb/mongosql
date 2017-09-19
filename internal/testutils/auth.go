package testutils

import (
	"os"

	toolsoptions "github.com/mongodb/mongo-tools/common/options"
)

func SqldTestAuthOpts() *toolsoptions.Auth {
	return &toolsoptions.Auth{
		Username: "bob",
		Password: "pwd123",
	}
}

func getAuthOpts() *toolsoptions.Auth {
	authOpts := &toolsoptions.Auth{}
	if len(os.Getenv("SQLPROXY_AUTHTEST")) > 0 {
		return SqldTestAuthOpts()
	}
	return authOpts
}
