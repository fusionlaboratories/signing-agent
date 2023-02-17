package main

const PrivateKey = ``

const APIKey = ""

type PartnerTxStatus string

const (
	PartnerTxStatusUnknown  PartnerTxStatus = "unknown"
	PartnerTxStatusNew      PartnerTxStatus = "new"
	PartnerTxStatusExisting PartnerTxStatus = "existing"
)

type WalletType int

// Wallet types
const (
	WalletTypeStandard          WalletType = 0
	WalletTypeDedicated         WalletType = 1
	WalletTypeMultiCounterparty WalletType = 2
	WalletTypeVesting           WalletType = 3
	WalletTypeExternal          WalletType = 4
	WalletTypeFee               WalletType = 5
)

// EntityType -
// swagger:enum EntityType
type EntityType string

// EntityTypes
const (
	EntityTypeUser       EntityType = "user"
	EntityTypeCompany    EntityType = "Company"
	EntityTypePartnerKey EntityType = "partner-key"
	EntityTypeCoreClient EntityType = "core-client"
	EntityTypeServer     EntityType = "server"
)

type ConnectPlugin string

const (
	ConnectPluginNone            ConnectPlugin = ""
	ConnectPluginMMI             ConnectPlugin = "mmi"
	ConnectPluginWalletConnectV1 ConnectPlugin = "wcv1"
)

type ActionStatus string

const (
	ActionStatusStrUnknown ActionStatus = "unknown"
	ActionStatusPending    ActionStatus = "pending"
	ActionStatusExpired    ActionStatus = "expired"
	ActionStatusApproved   ActionStatus = "approved"
	ActionStatusRejected   ActionStatus = "rejected"
)
