module github.com/10gen/sqlproxy

go 1.12

require (
	github.com/10gen/candiedyaml v0.0.0-20190528154413-da6f4db266e5
	github.com/10gen/mongoast v0.0.0-20191025180923-3d290796045c
	github.com/10gen/openssl v0.0.0-20180319163546-426ceace7847
	github.com/apmckinlay/gsuneido v0.0.0-20190828171039-d059fc31c5ab // indirect
	github.com/aws/aws-sdk-go v1.23.10 // indirect
	github.com/craiggwilson/goke v0.0.0-20190811215818-c46ad960904c // indirect
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/go-cmp v0.3.1
	github.com/gopherjs/gopherjs v0.0.0-20170702153443-2b1d432c8a82 // indirect
	github.com/howeyc/gopass v0.0.0-20170109162249-bf9dde6d0d2c
	github.com/jessevdk/go-flags v0.0.0-20170722072952-6cf8f02b4ae8
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/kardianos/osext v0.0.0-20170510131534-ae77be60afb1 // indirect
	github.com/kardianos/service v0.0.0-20180302231109-0ab6efe2ea51
	github.com/kr/pretty v0.0.0-20160823170715-cfb55aafdaf3
	github.com/kr/text v0.0.0-20160504234017-7cafcd837844 // indirect
	github.com/lib/pq v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/mattn/go-isatty v0.0.9 // indirect
	github.com/mongodb/mongo-tools v0.0.0-20191008165040-976b41822808
	github.com/mongodb/mongo-tools-common v1.0.2
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/pkg/profile v1.3.0 // indirect
	github.com/satori/go.uuid v0.0.0-20181028125025-b2ce2384e17b
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24
	github.com/smartystreets/assertions v0.0.0-20160225170624-e9a2f6771d97 // indirect
	github.com/smartystreets/goconvey v0.0.0-20160503033757-d4c757aa9afd
	github.com/spacemonkeygo/spacelog v0.0.0-20170706210657-b6d9bf7bf3eb // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	github.com/tidwall/pretty v0.0.0-20190325153808-1166b9ac2b65 // indirect
	github.com/tkuchiki/go-timezone v0.1.4
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v0.0.0-20180714160509-73f8eece6fdc
	go.mongodb.org/mongo-driver v1.1.2
	golang.org/x/crypto v0.0.0-20190820162420-60c769a6c586
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297 // indirect
	golang.org/x/sync v0.0.0-20190423024810-112230192c58 // indirect
	golang.org/x/sys v0.0.0-20190826190057-c7b8b68b1456 // indirect
	golang.org/x/text v0.3.2
	google.golang.org/appengine v1.6.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.2.1
)

replace (
	github.com/10gen/candiedyaml => github.com/10gen/candiedyaml v0.0.0-20190528154413-da6f4db266e5
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
	golang.org/x/text => golang.org/x/text v0.3.0
	gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.1
)
