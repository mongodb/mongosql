module github.com/10gen/sqlproxy

go 1.12

require (
	github.com/10gen/candiedyaml v0.0.0-20200803184741-f53dbc0225e1
	github.com/10gen/mongoast v0.0.0-20200709194530-675a03daa881
	github.com/10gen/openssl v0.0.0-20180319163546-426ceace7847
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/go-cmp v0.5.2
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/jessevdk/go-flags v1.4.0
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/kardianos/osext v0.0.0-20170510131534-ae77be60afb1 // indirect
	github.com/kardianos/service v0.0.0-20180302231109-0ab6efe2ea51
	github.com/kr/pretty v0.1.0
	github.com/kr/text v0.0.0-20160504234017-7cafcd837844 // indirect
	github.com/mongodb/mongo-tools v0.0.0-20210318165052-4b84777b8f84
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v0.0.0-20181028125025-b2ce2384e17b
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24
	github.com/smartystreets/assertions v0.0.0-20160225170624-e9a2f6771d97 // indirect
	github.com/smartystreets/goconvey v1.6.4
	github.com/spacemonkeygo/spacelog v0.0.0-20170706210657-b6d9bf7bf3eb // indirect
	github.com/stretchr/testify v1.6.1
	github.com/tkuchiki/go-timezone v0.1.4
	github.com/xdg/stringprep v1.0.1-0.20180714160509-73f8eece6fdc
	go.mongodb.org/mongo-driver v1.5.0
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073
	golang.org/x/text v0.3.3
	google.golang.org/appengine v1.6.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace (
	github.com/10gen/openssl => github.com/10gen/openssl v0.0.0-20180319163546-426ceace7847
	github.com/go-sql-driver/mysql => github.com/go-sql-driver/mysql v1.4.1
	github.com/google/go-cmp => github.com/google/go-cmp v0.2.0
	github.com/howeyc/gopass => github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/jessevdk/go-flags => github.com/jessevdk/go-flags v0.0.0-20170722072952-6cf8f02b4ae8
	github.com/kardianos/service => github.com/kardianos/service v0.0.0-20180302231109-0ab6efe2ea51
	github.com/kr/pretty => github.com/kr/pretty v0.0.0-20160823170715-cfb55aafdaf3
	github.com/pkg/errors => github.com/pkg/errors v0.8.1
	github.com/satori/go.uuid => github.com/satori/go.uuid v0.0.0-20181028125025-b2ce2384e17b
	github.com/shopspring/decimal => github.com/shopspring/decimal v0.0.0-20160918205201-d6f52241f332
	github.com/smartystreets/goconvey => github.com/smartystreets/goconvey v0.0.0-20160503033757-d4c757aa9afd
	github.com/stretchr/testify => github.com/stretchr/testify v1.3.0
	github.com/xdg/stringprep => github.com/xdg/stringprep v0.0.0-20180714160509-73f8eece6fdc
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.1
)
