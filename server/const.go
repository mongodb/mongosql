package server

const (
	// DefaultServerPort is the port on which the server runs by default.
	DefaultServerPort string = "3307"
)

// This indicates that the server or client is using the 4.1 protocol.
const clientProtocol41 = 0x00000200

const (
	minProtocolVersion                byte = 10
	maxPayloadLength                  int  = (1 << 24) - 1
	clearPasswordClientAuthPluginName      = "mysql_clear_password"
	nativePasswordPluginName               = "mysql_native_password"
	mongosqlAuthClientAuthPluginName       = "mongosql_auth"
)

// MySQL packet headers
const (
	OkHeader          byte = 0x00
	ErrHeader         byte = 0xff
	EOFHeader         byte = 0xfe
	LocalInFileHeader byte = 0xfb
)

// Server status constants
const (
	ServerStatusInTrans            uint16 = 0x0001
	ServerSatusAutocommit          uint16 = 0x0002
	ServerMoreResultsExists        uint16 = 0x0008
	ServerStatusNoGoodIndexUsed    uint16 = 0x0010
	ServerStatusNoIndexUsed        uint16 = 0x0020
	ServerStatusCursorExists       uint16 = 0x0040
	ServerStatusLastRowSend        uint16 = 0x0080
	ServerStatusDBDropped          uint16 = 0x0100
	ServerStatusNoBackslashEscaped uint16 = 0x0200
	ServerStatusMetadataChanged    uint16 = 0x0400
	ServerQueryWasSlow             uint16 = 0x0800
	ServerPsOutParams              uint16 = 0x1000
)

// Command constants
const (
	ComSleep byte = iota
	ComQuit
	ComInitDB
	ComQuery
	ComFieldList
	ComCreateDB
	ComDropDB
	ComRefresh
	ComShutdown
	ComStatistics
	ComProcessInfo
	ComConnect
	ComProcessKill
	ComDebug
	ComPing
	ComTime
	ComDelayedInsert
	ComChangeUser
	ComBinlogDump
	ComTableDump
	ComConnectOut
	ComRegisterSlave
	ComStmtPrepare
	ComStmtExecute
	ComStmtSendLongData
	ComStmtClose
	ComStmtReset
	ComSetOption
	ComStmtFetch
	ComDaemon
	ComBinlogDumpGtid
	ComResetConnection
)

// Client setting constants
const (
	ClientLongPassword uint32 = 1 << iota
	ClientFoundRows
	ClientLongFlag
	ClientConnectWithDB
	ClientNoSchema
	ClientCompress
	ClientODBC
	ClientLocalFiles
	ClientIgnoreSpace
	ClientProtocol41
	ClientInteractive
	ClientSSL
	ClientIgnoreSigpipe
	ClientTransactions
	ClientReserved
	ClientSecureConnection
	ClientMultiStatements
	ClientMultiResults
	ClientPsMultiResults
	ClientPluginAuth
	ClientConnectAttrs
	ClientPluginAuthLenencClientData
)

// MySQL type constants
const (
	MySQLTypeDecimal byte = iota
	MySQLTypeTiny
	MySQLTypeShort
	MySQLTypeLong
	MySQLTypeFloat
	MySQLTypeDouble
	MySQLTypeNull
	MySQLTypeTimestamp
	MySQLTypeLongLong
	MySQLTypeInt24
	MySQLTypeDate
	MySQLTypeTime
	MySQLTypeDatetime
	MySQLTypeYear
	MySQLTypeNewDate
	MySQLTypeVarchar
	MySQLTypeBit
)

// MySQL type constants specific to types in BSON without
// direct MySQL correspondence
const (
	MySQLTypeNewDecimal byte = iota + 0xf6
	MySQLTypeEnum
	MySQLTypeSet
	MySQLTypeTinyBlob
	MySQLTypeMediumBlob
	MySQLTypeLongBlob
	MySQLTypeBlob
	MySQLTypeVarString
	MySQLTypeString
	MySQLTypeGeometry
)

// 1 bit flags for MySQL column attributes
const (
	NotNullFlag       = 1
	PriKeyFlag        = 2
	UniqueKeyFlag     = 4
	BlobFlag          = 16
	UnsignedFlag      = 32
	ZerofillFlag      = 64
	BinaryFlag        = 128
	EnumFlag          = 256
	AutoIncrementFlag = 512
	TimestampFlag     = 1024
	SetFlag           = 2048
	NumFlag           = 32768
	PartKeyFlag       = 16384
	GroupFlag         = 32768
	UniqueFlag        = 65536
)
