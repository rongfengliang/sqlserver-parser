package packettype

// const define for tds operator
const (
	SQL_BATCH           = 0x01
	RPC_REQUEST         = 0x03
	TABULAR_RESULT      = 0x04
	ATTENTION           = 0x06
	BULK_LOAD           = 0x07
	TRANSACTION_MANAGER = 0x0E
	LOGIN7              = 0x10
	NTLMAUTH_PKT        = 0x11
	PRELOGIN            = 0x12
	FEDAUTH_TOKEN       = 0x08
)
