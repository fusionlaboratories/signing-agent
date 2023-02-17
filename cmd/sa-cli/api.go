package main

type ConnectAccount struct {
	CounterpartyID string `json:"counterparty_id"`
	ClientID       string `json:"client_id"`
}

type CustodyGroup struct {
	Threshold int      `json:"threshold"`
	Members   []string `json:"members"`
}

type AddFundRequest struct {
	FundName             string       `json:"name"`
	FundDescription      string       `json:"description"`
	CustodygroupWithdraw CustodyGroup `json:"custodygroup_withdraw"`
	CustodygroupTx       CustodyGroup `json:"custodygroup_tx"`
	Assets               []string     `json:"assets"`
	Members              []string     `json:"members"`
	Wallets              []FundWallet `json:"wallets"`
}

type FundWallet struct {
	Name                 string            `json:"name"`
	Asset                string            `json:"asset"`
	CustodygroupWithdraw *CustodyGroup     `json:"custodygroup_withdraw,omitempty"`
	CustodygroupTx       *CustodyGroup     `json:"custodygroup_tx,omitempty"`
	Type                 WalletType        `json:"type"`
	ConnectAccounts      *[]ConnectAccount `json:"connect,omitempty"`
}

type AddFundResponse struct {
	FundID               string `json:"fund_id"`
	CustodygroupWithdraw string `json:"custodygroup_withdraw"`
	CustodygroupTx       string `json:"custodygroup_tx"`
}

type AddWhitelistRequest struct {
	Address string `json:"address"`
	Asset   string `json:"asset"`
	Name    string `json:"name"`
}

type AssetAmount struct {
	Asset  string `json:"symbol"`
	Amount int64  `json:"amount"`
}

type NewWithdrawRequest struct {
	FundID      string      `json:"fund_id"`
	WalletID    string      `json:"wallet_id"`
	Address     string      `json:"address"`
	Send        AssetAmount `json:"send"`
	Reference   string      `json:"reference"`
	BenefitOf   string      `json:"benefit_of"`
	AccountNo   string      `json:"account_no"`
	Expires     int64       `json:"expires"`
	PartnerTxID string      `json:"partner_txID"`
}

type NewTransactionResponse struct {
	TxID   string          `json:"tx_id"`
	Status PartnerTxStatus `json:"status"`
}

type DepositAddress struct {
	Asset   string `json:"asset"`
	Address string `json:"address"`
	Balance int64  `json:"balance"`
}

type DepositAddressListResponse struct {
	TotalCount int               `json:"total_count"`
	List       []*DepositAddress `json:"list"`
}

type ClientRegisterInitRequest struct {
	Name         string `json:"name"`
	BLSPublicKey string `json:"blsPublicKey"`
	ECPublicKey  string `json:"ecPublicKey"`
}

type ClientRegisterInitResponse struct {
	ID           string `json:"id"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	AccountCode  string `json:"accountCode"`
	IDDocument   string `json:"idDoc"`
	Timestamp    int64  `json:"timestamp"`
}

type ClientRegisterResponse struct {
	ClientRegisterInitResponse
	FeedURL string `json:"feed_url"`
}

type CompanyPermissions struct {
	IsAdmin      bool `json:"admin"`
	CanFundAdmin bool `json:"fundManager"`
	CanTrade     bool `json:"trader"`
	CanCustody   bool `json:"custodian"`
}

type CompanyMemberFundStats struct {
	FundManager int `json:"fundManagerCount"`
	Trader      int `json:"traderCount"`
	Custodian   int `json:"custodianCount"`
}

type EntityUser struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Username  string `json:"username"`
}

type EntityExternalCounterparty struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	LogoURL string `json:"logourl"`
}

type EntityCompany struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Domain  string `json:"domain"`
}

type EntityCoreClient struct {
	Name string `json:"name"`
}

type Entity struct {
	ID                   string                      `json:"id"`
	AccountCode          string                      `json:"accountCode"`
	Name                 string                      `json:"name"`
	Initials             string                      `json:"initials"`
	Type                 EntityType                  `json:"type"`
	User                 *EntityUser                 `json:"user,omitempty"`
	Company              *EntityCompany              `json:"Company,omitempty"`
	CoreClient           *EntityCoreClient           `json:"coreClient,omitempty"`
	ExternalCounterparty *EntityExternalCounterparty `json:"externalCounterparty,omitempty"`
}

type CompanyMember struct {
	UserID      string                  `json:"userID"`
	Permissions CompanyPermissions      `json:"permissions"`
	Entity      *Entity                 `json:"entity,omitempty"`
	Status      ActionStatus            `json:"status"`
	FundStats   *CompanyMemberFundStats `json:"fundStats,omitempty"`
}

type ExternalClient struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Asset    string        `json:"asset"`
	PluginID ConnectPlugin `json:"pluginID"`
}

type ExternalClients struct {
	Clients []ExternalClient `json:"clients"`
}

type Company struct {
	CompanyID    string           `json:"companyID"`
	Name         string           `json:"name"`
	Domain       string           `json:"domain"`
	Address1     string           `json:"address1"`
	Address2     string           `json:"address2"`
	City         string           `json:"city"`
	State        string           `json:"state"`
	ZIP          string           `json:"zip"`
	CountryCode  string           `json:"countryCode"`
	Country      string           `json:"country"`
	MembersCount int              `json:"membersCount"`
	Members      []CompanyMember  `json:"members"`
	External     *ExternalClients `json:"external,omitempty"`
}

type CreateCompanyRequest struct {
	Company
	Ref string `json:"ref"`
}

type CreateCompanyResponse struct {
	CompanyID string `json:"company_id"`
	Ref       string `json:"ref"`
}

type AddTrustedPartyRequest struct {
	Address string `json:"address"`
}

type httpRequest struct {
	url       string
	method    string
	body      []byte
	timestamp string
}
