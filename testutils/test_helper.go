// +build !ssl

package testutils

import (
	"github.com/10gen/sqlproxy/options"
	tooloptions "github.com/mongodb/mongo-tools/common/options"
)

func GetSSLOpts() *tooloptions.SSL {
	return &tooloptions.SSL{}
}

func GetDrdlSSLOpts() *options.DrdlSSL {
	return &options.DrdlSSL{}
}
