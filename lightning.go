package golight

import (
	"fmt"
	"reflect"
	"github.com/niftynei/golight/jrpc2"
	"os"
)

// This file's the one that holds all the objects for the 
// c-lightning RPC commands 

type Lightning struct {
	client *jrpc2.Client
}

func NewLightning() *Lightning {
	ln := &Lightning{}
	ln.client = jrpc2.NewClient()
	return ln
}

func (l *Lightning) StartUp(in, out *os.File) {
	l.client.StartUp(in, out)
}

func (l *Lightning) Shutdown() {
	l.client.Shutdown()
}

type ListPeersRequest struct {
	PeerId string	`json:"id,omitempty"`
	Level string	`json:"level,omitempty"`
}

func (r *ListPeersRequest) Name() string {
	return "listpeers"
}

type PeersResponse struct {
	Peers []Peer	`json:"peers"`
}

type Peer struct {
	Id string	`json:"id"`
	Connected bool	`json:"connected"`
	NetAddresses []string	`json:"netaddr"`
	GlobalFeatures string	`json:"globalfeatures"`
	LocalFeatures string	`json:"localfeatures"`
	Channels []PeerChannel	`json:"channels"`
	Logs []Log	`json:"log,omitempty"`
}

type PeerChannel struct {
	State string	`json:"state"`
	ScratchTxId string	`json:"scratch_txid"`
	Owner string	`json:"owner"`
	ShortChannelId string	`json:"short_channel_id"`
	ChannelId string	`json:"channel_id"`
	FundingTxId string	`json:"funding_txid"`
	Funding string	`json:"funding"`
	Status []string	`json:"status"`
	MilliSatoshiToUs uint64	`json:"msatoshi_to_us"`
	MilliSatoshiToUsMin uint64	`json:"msatoshi_to_us_min"`
	MilliSatoshiToUsMax uint64	`json:"msatoshi_to_us_max"`
	MilliSatoshiTotal uint64	`json:"msatoshi_total"`
	DustLimitSatoshi uint64		`json:"dust_limit_satoshis"`
	MaxHtlcValueInFlightMilliSatoshi uint64	`json:"max_htlc_value_in_flight_msat"`
	TheirChannelReserveSatoshi uint64	`json:"their_channel_reserve_satoshis"`
	OurChannelReserveSatoshi uint64	`json:"our_channel_reserve_satoshis"`
	SpendableMilliSatoshi uint64	`json:"spendable_msatoshis"`
	HtlcMinMilliSatoshi uint64	`json:"htlc_minimum_msat"`
	TheirToSelfDelay uint	`json:"their_to_self_delay"`
	OurToSelfDelay uint	`json:"our_to_self_delay"`
	MaxAcceptedHtlcs uint	`json:"max_accepted_htlcs"`
	InPaymentsOffered uint64	`json:"in_payments_offered"`
	InMilliSatoshiOffered uint64	`json:"in_msatoshi_offered"`
	InPaymentsFulfilled uint64	`json:"in_payments_fulfilled"`
	InMilliSatoshiFulfilled uint64	`json:"in_msatoshi_fulfilled"`
	OutPaymentsOffered uint64	`json:"out_payments_offered"`
	OutMilliSatoshiOffered uint64	`json:"out_msatoshi_offered"`
	OutPaymentsFulfilled uint64	`json:"out_payments_fulfilled"`
	OutMilliSatoshiFulfilled uint64	`json:"out_msatoshi_fulfilled"`
	Htlcs []Htlc	`json:"htlcs"`
}

type Htlc struct {
	Direction string	`json:"direction"`
	Id uint64	`json:"id"`
	MilliSatoshi uint64	`json:"msatoshi"`
	Expiry uint64	`json:"expiry"`
	PaymentHash string	`json:"payment_hash"`
	State string	`json:"state"`
}

// Show current peer {peerId}. If {level} is set, include logs.
func (l *Lightning) GetPeer(peerId string, level LogLevel) (*PeersResponse, error) {
	var result PeersResponse
	err := l.client.Request(&ListPeersRequest{peerId,level.String()}, result)
	return &result, err
}

// Show current peers, if {level} is set, include logs.
func (l *Lightning) ListPeers(level LogLevel) (*PeersResponse, error) {
	return l.GetPeer("", level)
}


type ListNodeRequest struct {
	NodeId	string	`json:"id,omitempty"`
}

func (ln *ListNodeRequest) Name() string {
	return "listnodes"
}

type Nodes struct {
	Nodes []Node	`json:"nodes"`
}

type Node struct {
	Id string	`json:"nodeid"`
	Alias string	`json:"alias"`
	Color string	`json:"color"`
	LastTimestamp uint	`json:"last_timestamp"`
	GlobalFeatures string	`json:"globalfeatures"`
	Addresses []Address	`json:"addresses"`
}

type Address struct {
	// todo: map to enum (ipv4, ipv6, torv2, torv3)
	Type string	`json:"type"`
	Address string	`json:"address"`
	Port int	`json:"port"`
}

// Get all nodes in our local network view, filter on node {id},
// if provided
func (l *Lightning) GetNode(nodeId string) (*Nodes, error) {
	var result Nodes
	err := l.client.Request(&ListNodeRequest{nodeId}, &result)
	return &result, err
}

// List all nodes in our local network view
func (l *Lightning) ListNodes() (*Nodes, error) {
	return l.GetNode("")
}

type RouteRequest struct {
	PeerId string		`json:"id"`
	MilliSatoshis uint64	`json:"msatoshi"`
	RiskFactor float32	`json:"riskfactor"`
	Cltv	uint		`json:"cltv"`
	FromId	string		`json:"fromid,omitempty"`
	FuzzPercent float32	`json:"fuzzpercent"`
	Seed	string		`json:"seed,omitempty'`
}

type Route struct {
	Route []RouteHop	`json:"route"`
}

type RouteHop struct {
	Id string	`json:"id"`
	ShortChannelId string	`json:"channel"`
	MilliSatoshi uint64	`json:"msatoshi"`
	Delay uint	`json:"delay"`
}

func (rr *RouteRequest) Name() string {
	return "getroute"
}

// Show route to {id} for {msatoshis}, using a {riskfactor} and optional
// {cltv} value (defaults to 9). If specified, search from {fromId} otherwise
// use current node as the source. Randomize the route with up to {fuzzpercent}
// (0.0 -> 100.0, default 5.0) using {seed} as an arbitrary-size string seed.
func (l *Lightning) GetRoute(peerId string, msats uint64, riskfactor float32, cltv uint, fromId string, fuzzpercent float32, seed string) (*Route, error) {
	if peerId == "" {
		return nil, fmt.Errorf("Must provide a peerId to route to")
	}

	if msats == 0 {
		return nil, fmt.Errorf("No value set for payment. (`msatoshis` is equal to zero).")
	}

	if riskfactor <= 0 || riskfactor >= 100 {
		return nil, fmt.Errorf("The risk factor must set above 0 and beneath 100")
	}

	if fuzzpercent == 0 {
		fuzzpercent = 5.0
	} else if fuzzpercent < 0 || fuzzpercent > 100 {
		return nil, fmt.Errorf("The `fuzzpercent` value must be between 0 and 100")
	}

	if cltv == 0 {
		cltv = 9
	}

	var result Route
	err := l.client.Request(&RouteRequest{
		PeerId: peerId,
		MilliSatoshis: msats,
		RiskFactor: riskfactor,
		Cltv: cltv,
		FromId: fromId,
		FuzzPercent: fuzzpercent,
		Seed: seed,
	}, &result)
	return &result, err
}

type ListChannelRequest struct {
	ShortChannelId string	`json:"short_channel_id"`
}

func (lc *ListChannelRequest) Name() string {
	return "listchannels"
}

type ChannelResponse struct {
	Channels []Channel	`json:"channels"`
}

type Channel struct {
	Source string	`json:"source"`
	Destination string	`json:"destination"`
	ShortChannelId string	`json:"short_channel_id"`
	IsPublic bool	`json:"public"`
	Satoshis uint64	`json:"satoshis"`
	MessageFlags uint	`json:"message_flags"`
	ChannelFlags uint	`json:"channel_flags"`
	IsActive bool	`json:"active"`
	LastUpdate uint	`json:"last_update"`
	BaseFeeMillisatoshi uint64	`json:"base_fee_millisatoshi"`
	FeePerMillionth float64	`json:"fee_per_millionth"`
	Delay uint	`json:"delay"`
}

// Get channel by {shortChanId}
func (l *Lightning) GetChannel(shortChanId string) (*ChannelResponse, error) {
	var result ChannelResponse
	err := l.client.Request(&ListChannelRequest{shortChanId}, &result)
	return &result, err
}

func (l *Lightning) ListChannels() (interface{}, error) {
	var result interface{}
	err := l.client.Request(&ListChannelRequest{}, result)
	return result, err
}

type InvoiceRequest struct {
	MilliSatoshis string	`json:"msatoshi"`
	Label string	`json:"label"`
	Description string	`json:"description"`
	ExpirySeconds uint32	`json:"expiry,omitempty"`
	Fallbacks []string	`json:"fallbacks,omitempty"`
	PreImage string	`json:"preimage,omitempty"`
}

func (ir *InvoiceRequest) Name() string {
	return "invoice"
}

type Invoice struct {
	PaymentHash string	`json:"payment_hash"`
	ExpiresAt uint64	`json:"expires_at"`
	Bolt11 string	`json:"bolt11"`
	WarningOffline string	`json:"warning_offline"`
	WarningCapacity string	`json:"warning_capacity"`
}

// Creates an invoice with a value of "any", that can be paid with any amount
func (l *Lightning) CreateInvoiceAny(label, description string, expirySeconds uint32, fallbacks []string, preimage string) (*Invoice, error) {
	return createInvoice(l, "any", label, description, expirySeconds, fallbacks, preimage)
}

// Creates an invoice with a value of `msat`. Label and description must be set.
//
// The 'label' is a unique string or number (which is treated as a string); it is
// never revealed to other nodes, but it can be used to query the status of this 
// invoice.
//
// The 'description' is a short description of purpose of payment. It is encoded 
// into the invoice. Must be UTF-8, cannot use '\n' JSON escape codes.
//
// The 'expiry' is optionally the number of seconds the invoice is valid for.
// Defaults to 3600 (1 hour).
// 
// 'fallbacks' is one or more fallback addresses to include in the invoice. They
// should be ordered from most preferred to least. Noe that these are not
// currently tracked to fulfill the invoice.
//
// The 'preimage' is a 64-digit hex string to be used as payment preimage for 
// the created invoice. By default, c-lightning will generate a secure 
// pseudorandom preimage seeded from an appropriate entropy source on your 
// system. **NOTE**: if you specify the 'preimage', you are responsible for 
// both ensuring that a suitable psuedorandom generator with sufficient entropy 
// was used in its creation and keeping it secret. 
// This parameter is an advanced feature intended for use with cutting-edge 
// cryptographic protocols and should not be used unless explicitly needed.
func (l *Lightning) CreateInvoice(msat uint64, label, description string, expirySeconds uint32, fallbacks []string, preimage string) (*Invoice, error) {

	if msat <= 0 {
		return nil, fmt.Errorf("No value set for invoice. (`msat` is less than or equal to zero).")
	}
	return createInvoice(l, string(msat), label, description, expirySeconds, fallbacks, preimage)

}

func createInvoice(l *Lightning, msat, label, description string, expirySeconds uint32, fallbacks []string, preimage string) (*Invoice, error) {

	if label == "" {
		return nil, fmt.Errorf("Must set a label on an invoice")
	}
	if description == "" {
		return nil, fmt.Errorf("Must set a description on an invoice")
	}

	var result Invoice
	err := l.client.Request(&InvoiceRequest{
		MilliSatoshis: msat,
		Label: label,
		Description: description,
		ExpirySeconds: expirySeconds,
		Fallbacks: fallbacks,
		PreImage: preimage,
	}, &result)
	return &result, err
}

type ListInvoiceRequest struct {
	Label string	`json:"label,omitempty"`
}

func (r *ListInvoiceRequest) Name() string {
	return "listinvoices"
}

// List all invoices
func (l *Lightning) ListInvoices() ([]Invoice, error) {
	return l.GetInvoice("")
}

// Show invoice {label}.
func (l *Lightning) GetInvoice(label string) ([]Invoice, error) {
	var result struct {
		List []Invoice `json:"invoices"`
	}
	err := l.client.Request(&ListInvoiceRequest{label}, &result)
	return result.List, err
}

type DeleteInvoiceRequest struct {
	Label string	`json:"label"`
	Status string	`json:"status"`
}

func (r *DeleteInvoiceRequest) Name() string {
	return "delinvoice"
}

// Delete unpaid invoice {label} with {status}
func (l *Lightning) DeleteInvoice(label, status string) (*Invoice, error) {
	var result Invoice
	err := l.client.Request(&DeleteInvoiceRequest{label,status}, &result)
	return &result, err
}

type WaitAnyInvoiceRequest struct {
	LastPayIndex uint	`json:"lastpay_index"`
}

func (r *WaitAnyInvoiceRequest) Name() string {
	return "waitanyinvoice"
}

// Waits until an invoice is paid, then returns a single entry. 
// Will not return or provide any invoices paid prior to or including 
// the lastPayIndex.
// 
// The 'pay index' is a monotonically-increasing number assigned to
// an invoice when it gets paid. The first valid 'pay index' is 1.
func (l *Lightning) WaitAnyInvoice(lastPayIndex uint) (interface{}, error) {
	var result interface{}
	err := l.client.Request(&WaitAnyInvoiceRequest{lastPayIndex}, result)
	return result, err
}

type WaitInvoiceRequest struct {
	Label string	`json:"label"`
}

func (r *WaitInvoiceRequest) Name() string {
	return "waitinvoice"
}

func (l *Lightning) WaitInvoice(label string) (*Invoice, error) {
	if label == "" {
		return nil, fmt.Errorf("Must call wait invoice with a label")
	}

	var result Invoice
	err := l.client.Request(&WaitInvoiceRequest{label}, &result)
	return &result, err
}

type DecodePayRequest struct {
	Bolt11 string	`json:"bolt11"`
	Description string	`json:"description,omitempty"`
}

func (r *DecodePayRequest) Name() string {
	return "decodepay"
}

type DecodedBolt11 struct {
	Currency string	`json:"currency"`
	CreatedAt uint64	`json:"created_at"`
	Expiry uint64	`json:"expiry"`
	Payee string	`json:"payee"`
	MilliSatoshis uint64	`json:"msatoshi"`
	Description string	`json:"description"`
	DescriptionHash string	`json:"description_hash"`
	MinFinalCltvExpiry float64	`json:"min_final_cltv_expiry"`
	Fallbacks []Fallback	`json:"fallbacks"`
	Routes []BoltRoute	`json:"routes"`
	Extra []BoltExtra	`json:"extra"`
	PaymentHash string	`json:"payment_hash"`
	Signature string	`json:"signature"`
}

type Fallback struct {
	// fixme: use enum (P2PKH,P2SH,P2WPKH,P2WSH)
	Type string	`json:"type"`
	Address string	`json:"addr"`
	Hex string	`json:"hex"`
}

type BoltRoute struct {
	Pubkey string	`json:"pubkey"`
	ShortChannelId string	`json:"short_channel_id"`
	FeeBaseMilliSatoshis uint64	`json:"fee_base_msat"`
	FeeProportionalMillionths uint64	`json:"fee_proportional_millionths"`
	CltvExpiryDelta uint	`json:"cltv_expiry_delta"`
}

type BoltExtra struct {
	Tag string	`json:"tag"`
	Data string	`json:"data"`
}

// Decode the {bolt11}, using the provided 'description' if necessary.*
//
// * This is only necesary if the bolt11 includes a description hash.
// The provided description must match the included hash.
func (l *Lightning) DecodePay(bolt11, desc string) (*DecodedBolt11, error) {
	if bolt11 == "" {
		return nil, fmt.Errorf("Must call decode pay with a bolt11")
	}

	var result DecodedBolt11
	err := l.client.Request(&DecodePayRequest{bolt11, desc}, &result)
	return &result, err
}

type HelpRequest struct {}

func (r *HelpRequest) Name() string {
	return "help"
}

type Command struct {
	NameAndUsage  string	`json:"command"`
	Description string	`json:"description"`
	Verbose string	`json:"verbose"`
}

// Show available c-lightning RPC commands
func (l *Lightning) Help() ([]Command, error) {
	var result struct {
		Commands []Command	`json:"help"`
	}
	err := l.client.Request(&HelpRequest{}, &result)
	return result.Commands, err
}

type StopRequest struct {}

func (r *StopRequest) Name() string {
	return "stop"
}

// Shut down the c-lightning process. Will return a string
// of "Shutting down" on success.
func (l *Lightning) Stop() (string, error) {
	var result string
	err := l.client.Request(&StopRequest{}, &result)
	return result, err
}

type LogLevel int

const (
	Info LogLevel = iota
	Unusual
	Debug
	Io
)

func (l LogLevel) String() string {
	return []string{
		"info",
		"unusual",
		"debug",
		"io",
	}[l]
}

type LogRequest struct {
	Level string	`json:"level,omitempty"`
}

func (r *LogRequest) Name() string {
	return "getlog"
}

type LogResponse struct {
	CreatedAt string	`json:"created_at"`
	BytesUsed uint64	`json:"bytes_used"`
	BytesMax uint64		`json:"bytes_max"`
	Logs []Log		`json:"log"`
}

type Log struct {
	Type string	`json:"type"`
	Time string	`json:"time,omitempty"`
	Source string	`json:"source,omitempty"`
	Message string	`json:"log,omitempty"`
	NumSkipped uint	`json:"num_skipped,omitempty"`
}

// Show logs, with optional log {level} (info|unusual|debug|io)
func (l *Lightning) GetLog(level LogLevel) (*LogResponse, error) {
	var result LogResponse
	err := l.client.Request(&LogRequest{level.String()}, &result)
	return &result, err
}

type DevRHashRequest struct {
	Secret string	`json:"secret"`
}

func (r *DevRHashRequest) Name() string {
	return "dev-rhash"
}

type DevHashResult struct {
	RHash string	`json:"rhash"`
}

// Show SHA256 of {secret}
func (l *Lightning) DevHash(secret string) (*DevHashResult, error) {
	if secret == "" {
		return nil, fmt.Errorf("Must pass in a valid secret to hash")
	}

	var result DevHashResult
	err := l.client.Request(&DevRHashRequest{secret}, &result)
	return &result, err
}

type DevCrashRequest struct {}

func (r *DevCrashRequest) Name() string {
	return "dev-crash"
}

// Crash lightningd by calling fatal(). Returns nothing.
func (l *Lightning) DevCrash() (interface{}, error) {
	err := l.client.Request(&DevCrashRequest{}, nil)
	return nil, err
}

type DevQueryShortChanIdsRequest struct {
	PeerId string	`json:"id"`
	ShortChanIds []string	`json:"scids"`
}

func (r *DevQueryShortChanIdsRequest) Name() string {
	return "dev-query-scids"
}

type QueryShortChannelIdsResponse struct {
	IsComplete bool	`json:"complete"`
}

// Ask a peer for a particular set of short channel ids
func (l *Lightning) DevQueryShortChanIds(peerId string, shortChanIds []string) (*QueryShortChannelIdsResponse, error) {
	if peerId == "" {
		return nil, fmt.Errorf("Must provide a peer id")
	}

	if len(shortChanIds) == 0 {
		return nil, fmt.Errorf("Must specify short channel ids to query for")
	}

	var result QueryShortChannelIdsResponse
	err := l.client.Request(&DevQueryShortChanIdsRequest{peerId, shortChanIds}, &result)
	return &result, err
}

type GetInfoRequest struct {}

func (r *GetInfoRequest) Name() string {
	return "getinfo"
}

type NodeInfo struct {
	Id string	`json:"id"`
	Alias string	`json:"alias"`
	Color string	`json:"color"`
	PeerCount int	`json:"num_peers"`
	PendingChannelCount int	`json:"num_pending_channels"`
	ActiveChannelCount int	`json:"num_active_channels"`
	InactiveChannelCount int	`json:"num_inactive_channels"`
	Addresses []Address	`json:"address"`
	Binding []AddressInternal	`json:"binding"`
	Version	string	`json:"version"`
	Blockheight int	`json:"blockheight"`
	Network string	`json:"network"`
	FeesCollectedMilliSatoshis uint64	`json:"msatoshi_fees_collected"`
}

type AddressInternal struct {
	Address
	Socket string	`json:"socket"`
	Service Address	`json:"service"`
	Name string	`json:"name"`
}

func (l *Lightning) GetInfo() (*NodeInfo, error) {
	var result NodeInfo
	err := l.client.Request(&GetInfoRequest{}, &result)
	return &result, err
}

type SendPayRequest struct {
	Route interface{} `json:"route"`
	PaymentHash string `json:"payment_hash"`
	Desc string	`json:"description,omitempty"`
	MilliSatoshis uint64 `json:"msatoshi,omitempty"`
}

func (r *SendPayRequest) Name() string {
	return "sendpay"
}

type PaymentFields struct {
	Id uint64	`json:"id"`
	PaymentHash string	`json:"payment_hash"`
	Destination string	`json:"destination"`
	MilliSatoshi uint64	`json:"msatoshi"`
	MilliSatoshiSent uint64	`json:"msatoshi_sent"`
	CreatedAt uint64	`json:"created_at"`
	Status string	`json:"status"`
	PaymentPreimage string	`json:"payment_preimage"`
	Description string	`json:"description"`
}

type SendPayResult struct {
	Message string	`json:"message"`
	PaymentFields
}

// Send along {route} in return for preimage of {paymentHash}
//  Description and msat are optional.
// Generally a client would call GetRoute to resolve a route, then 
// use SendPay to send it.  If it fails, it would call GetRoute again
// to retry.
//
// Response will occur when payment is on its way to the destination.
// Does not wati for a definitive success or failure. Use 'waitsendpay'
// to poll or wait for definite success or failure.
//
// 'description', if provided, will be returned in 'waitsendpay' and
// 'listpayments' results.
//
// 'msat', if provided, is the amount that will be recorded as the target
// payment value. If not specified, it will be the final amount to the 
// destination (specified in route).  If specified, then the final amount
// at the destination must be from the specified 'msat' to twice that
// value, inclusive. This is inteded to obscure payments by overpaying
// slightly at the destination -- the acutal target paymnt is what 
// should be specified as the 'msat' argument.
// 
// Once a payment has succeeded, calls to 'SendPay' with the same 
// 'paymentHash' but a different 'msat' or destination will fail; this
// prevents accidental multiple payments. Calls with the same 'paymentHash',
// 'msat' and destination as a previous successful payment will return
// immediately with a success, even if the route is different.
func (l *Lightning) SendPay(route interface{}, paymentHash, description string, msat uint64) (*SendPayResult, error) {
	if paymentHash == "" {
		return nil, fmt.Errorf("Must specify a paymentHash to pay")
	}
	// todo: an actual 'route' object specification
	if reflect.ValueOf(route).IsNil() {
		return nil, fmt.Errorf("Must specify a route to send payment along")
	}

	var result SendPayResult
	err := l.client.Request(&SendPayRequest{
		Route: route,
		PaymentHash: paymentHash,
		Desc: description,
		MilliSatoshis: msat,
	}, &result)
	return &result, err
}

type WaitSendPayRequest struct {
	PaymentHash string	`json:"payment_hash"`
	Timeout uint		`json:"timeout"`
}

func (r *WaitSendPayRequest) Name() string {
	return "waitsendpay"
}

// Polls or waits for the status of an outgoing payment that was 
// initiated by a previous 'SendPay' invocation.
//
// May provide a 'timeout, in seconds. When provided, will return a
// 200 error code (payment still in progress) if timeout elapses
// before the payment is definitively concluded (success or fail).
// If no 'timeout' is provided, the call waits indefinitely.
func (l *Lightning) WaitSendPay(paymentHash string, timeout uint) (*PaymentFields, error) {
	if paymentHash == "" {
		return nil, fmt.Errorf("Must provide a payment hash to pay")
	}

	var result PaymentFields
	err := l.client.Request(&WaitSendPayRequest{paymentHash, timeout}, &result)
	return &result, err

}

type PayRequest struct {
	Bolt11 string	`json:"bolt11"`
	MilliSatoshi uint64	`json:"msatoshi,omitempty"`
	Desc string	`json:"description,omitempty"`
	RiskFactor float32	`json:"riskfactor,omitempty"`
	MaxFeePercent float32	`json:"maxfeeprecent,omitempty"`
	RetryFor uint	`json:"retry_for,omitempty"`
	MaxDelay uint	`json:"maxdelay,omitempty"`
	ExemptFee bool	`json:"exemptfee,omitempty"`
}

func (r *PayRequest) Name() string {
	return "pay"
}

// todo: there's lots of different data that comes back for 
// payment failures, that for now we totally lose
type PaymentSuccess struct {
	PaymentFields
	GetRouteTries int	`json:"getroute_tries"`
	SendPayTries int	`json:"sendpay_tries"`
	Route []RouteHop	`json:"route"`
	Failures []PayFailures	`json:"failures"`
}

type PayFailures struct {
	Message string	`json:"message"`
	Type string	`json:"type"`
	OnionReply string	`json:"onionreply"`
	ErringIndex int	`json:"erring_index"`
	FailCode int	`json:"failcode"`
	ErringNode string	`json:"erring_node"`
	ErringShortChannelId string	`json:"erring_channel"`
	ChannelUpdate string	`json:"channel_update"`
	Route []RouteHop	`json:"route"`
}

func (l *Lightning) PayBolt(bolt11 string) (*PaymentSuccess, error) {
	return l.Pay(&PayRequest{
		Bolt11: bolt11,
	})
}

// Send payment as specified by 'Bolt11' with 'MilliSatoshi'
// (Millisatoshis amount is ignored if the 'Bolt11' includes an amount).
//
// 'description' is required if the 'bolt11' includes a description hash.
// 
// 'riskfactor' is optional, defaults to 1.0
// Briefly, the 'riskfactor' is the estimated annual cost of your funds
// being stuck (as a percentage), multiplied by the percent change of 
// each node failing. Ex: 1% chance of node failure and a 20% annual cost
// would give you a risk factor of 20. c-lightning defaults to 1.0
//
// 'MaxFeePercent' is the max percentage of a payment that can be paid 
// in fees. c-lightning defaults to 0.5.
// 
// 'ExemptFee' can be used for tiny paymetns which would otherwise be 
// dominated by the fee leveraged by forwarding nodes. Setting 'ExemptFee' 
// allows 'MaxFeePercent' check to be skipped on fees that are smaller than
// 'ExemptFee'. c-lightning default is 5000 millisatoshi.
//
// c-lightning will keep finding routes and retrying payment until it succeeds
// or the given 'RetryFor' seconds have elapsed.  Note that the command may 
// stop retrying while payment is pending. You can continuing monitoring
// payment status with the ListPayments or WaitSendPay. 'RetryFor' defaults
// to 60 seconds.
// 
// 'MaxDelay' is used when determining whether a route incurs an acceptable 
// delay. A route will not be used if the estimated delay is above this.
// Defaults to the configured locktime max (--max-locktime-blocks)
// Units is in blocks.
func (l *Lightning) Pay(req *PayRequest) (*PaymentSuccess, error) {
	if req.Bolt11 == "" {
		return nil, fmt.Errorf("Must supply a Bolt11 to pay")
	}
	if req.RiskFactor < 0 {
		return nil, fmt.Errorf("Risk factor must be postiive %f", req.RiskFactor)
	}
	if req.MaxFeePercent < 0 || req.MaxFeePercent > 100 {
		return nil, fmt.Errorf("MaxFeePercent must be a percentage. %f", req.MaxFeePercent)
	}
	var result PaymentSuccess
	err := l.client.Request(req, &result)
	return &result, err
}

type ListPaymentRequest struct {
	Bolt11 string	`json:"bolt11,omitempty"`
	PaymentHash string	`json:"payment_hash,omitempty"`
}

func (r *ListPaymentRequest) Name() string {
	return "listpayments"
}

// Show outgoing payments, regarding {bolt11}
func (l *Lightning) ListPayments(bolt11 string) ([]PaymentFields, error) {
	return l.listPayments(&ListPaymentRequest{
		Bolt11: bolt11,
	})
}

// Show outgoing payments, regarding {paymentHash}
func (l *Lightning) ListPaymentsHash(paymentHash string) ([]PaymentFields, error) {
	return l.listPayments(&ListPaymentRequest{
		PaymentHash: paymentHash,
	})
}

func (l *Lightning) listPayments(req *ListPaymentRequest) ([]PaymentFields, error) {
	var result struct {
		Payments []PaymentFields	`json:"payments"`
	}
	err := l.client.Request(req, &result)
	return result.Payments, err
}

type ConnectRequest struct {
	PeerId string	`json:"id"`
	Host string	`json:"host"`
	Port uint	`json:"port"`
}

func (r *ConnectRequest) Name() string {
	return "connect"
}

type ConnectSuccess struct {
	PeerId string	`json:"id"`
}

// Connect to {peerId} at {host}:{port}. Returns peer id on success
func (l *Lightning) Connect(peerId, host string, port uint) (string, error) {
	var result struct {
		Id string	`json:"id"`
	}
	err := l.client.Request(&ConnectRequest{peerId,host,port}, &result)
	return result.Id, err
}

type FundChannelRequest struct {
	Id string	`json:"id"`
	Satoshi uint64	`json:"satoshi"`
	FeeRate float32	`json:"feerate,omitempty"`
	Announce bool	`json:"announce,omitempty"`
}

func (r *FundChannelRequest) Name() string {
	return "fundchannel"
}

type FundChannelResult struct {
	FundingTx string	`json:"tx"`
	FundingTxId string	`json:"txid"`
	ChannelId string	`json:"channel_id"`
}

// Fund channel with node {id} using {satoshi} satoshis, with feerate of {feerate}. Uses
// default feerate if unset. 
// If announce is false, channel announcements will not be sent.
func (l *Lightning) FundChannel(id string, satoshis uint64, feerate float32, announce bool) (*FundChannelResult, error) {
	if feerate < 0 {
		return nil, fmt.Errorf("Feerate must be positive %f", feerate)
	}
	var result FundChannelResult
	err := l.client.Request(&FundChannelRequest{id,satoshis,feerate,announce}, &result)
	return &result, err
}

type CloseRequest struct {
	PeerId string	`json:"id"`
	Force bool	`json:"force,omitempty"`
	Timeout uint	`json:"timeout,omitempty"`
}

func (r *CloseRequest) Name() string {
	return "close"
}

// Close the channel with peer {id}, timing out with {timeout} seconds. 
// If unspecified, times out in 30 seconds. 
// 
// If {force} is set, and close attempt times out, the channel will be closed
// unilaterally from our side.
//
// Can pass either peer id or channel id as {id} field.
//
// Note that a successful result may be null.
func (l *Lightning) Close(id string, force bool, timeout uint) (interface{}, error) {
	var result interface{}
	err := l.client.Request(&CloseRequest{id, force, timeout}, result)
	return result, err
}

type DevSignLastTxRequest struct {
	PeerId string	`json:"id"`
}

func (r *DevSignLastTxRequest) Name() string {
	return "dev-sign-last-tx"
}

// Sign and show the last commitment transaction with peer {peerId}
func (l *Lightning) DevSignLastTx(peerId string) (interface{}, error) {
	var result interface{}
	err := l.client.Request(&DevSignLastTxRequest{peerId}, result)
	return result, err
}

type DevFailRequest struct {
	PeerId string	`json:"id"`
}

func (r *DevFailRequest) Name() string {
	return "dev-fail"
}

// Fail with peer {id}
func (l *Lightning) DevFail(peerId string) (interface{}, error) {
	var result interface{}
	err := l.client.Request(&DevFailRequest{peerId}, result)
	return result, err
}

type DevReenableCommitRequest struct {
	PeerId string	`json:"id"`
}

func (r *DevReenableCommitRequest) Name() string {
	return "dev-reenable-commit"
}

// Re-enable the commit timer on peer {id}
func (l *Lightning) DevReenableCommit(id string) (interface{}, error) {
	var result interface{}
	err := l.client.Request(&DevReenableCommitRequest{id}, result)
	return result, err
}

type PingRequest struct {
	Id string	`json:"id"`
	Len uint	`json:"len"`
	PongBytes uint	`json:"pongbytes"`
}

func (r *PingRequest) Name() string {
	return "ping"
}

// Send {peerId} a ping of size 128, asking for 128 bytes in response
func (l *Lightning) Ping(peerId string) (interface{}, error) {
	return l.PingWithLen(peerId, 128, 128)
}

// Send {peerId} a ping of length {pingLen} asking for bytes {pongByteLen}
func (l *Lightning) PingWithLen(peerId string, pingLen, pongByteLen uint) (interface{}, error) {
	var result interface{}
	err := l.client.Request(&PingRequest{peerId, pingLen, pongByteLen}, result)
	return result, err
}

type DevMemDumpRequest struct { }

func (r *DevMemDumpRequest) Name() string {
	return "dev-memdump"
}

// Show memory objects currently in use
func (l *Lightning) DevMemDump() (interface{}, error) {
	var result interface{}
	err := l.client.Request(&DevMemDumpRequest{}, result)
	return result, err
}

type DevMemLeakRequest struct {}

func (r *DevMemLeakRequest) Name() string {
	return "dev-memleak"
}

// Show unreferenced memory objects
func (l *Lightning) DevMemLeak() (interface{}, error) {
	var result interface{}
	err := l.client.Request(&DevMemLeakRequest{}, result)
	return result, err
}

type WithdrawRequest struct {
	Destination string	`json:"destination"`
	Satoshi string		`json:"satoshi"`
	FeeRate string		`json:"feerate,omitempty"`
}

type SatoshiAmount struct {
	Amount uint64
	SendAll bool
}

func (s *SatoshiAmount) String() string {
	if s.SendAll {
		return "all"
	}
	return string(s.Amount)
}

type FeeDirective int
const (
	Normal FeeDirective = iota
	Urgent
	Slow
)

func (f FeeDirective) String() string {
	return []string{
		"normal",
		"urgent",
		"slow",
	}[f]
}

type FeeRateStyle int

const (
	SatPerKiloByte FeeRateStyle = iota
	SatPerKiloSipa
)

type FeeRate struct {
	Rate uint
	Style FeeRateStyle
	Directive FeeDirective
}

func (r FeeRateStyle) String() string {
	return []string{"perkb","perkw"}[r]
}

func (f *FeeRate) String() string {
	if f.Rate > 0 {
		return string(f.Rate) + f.Style.String()
	}
	// defaults to 'normal'
	return f.Directive.String()
}

func (r *WithdrawRequest) Name() string {
	return "withdraw"
}

// Withdraw sends funds from c-lightning's internal wallet to the
// address specified in {destination}. Address can be of any Bitcoin
// accepted type, including bech32.
//
// {satoshi} is the amount to be withdrawn from the wallet.
// 
// {feerate} is an optional feerate to use. Can be either a directive
// (urgent, normal, or slow) or a number with an optional suffix.
// 'perkw' means the number is interpreted as satoshi-per-kilosipa (weight)
// and 'perkb' means it is interpreted bitcoind-style as satoshi-per-kilobyte.
// Omitting the suffix is equivalent to 'perkb'
// If not set, {feerate} defaults to 'normal'.
func (l *Lightning) Withdraw(destination string, amount *SatoshiAmount, feerate *FeeRate) (interface{}, error) {
	if amount == nil || (amount.Amount == 0 && !amount.SendAll) {
		return nil, fmt.Errorf("Must set satoshi amount to send")
	}
	request := &WithdrawRequest{
		Destination: destination,
		Satoshi: amount.String(),
	}
	if feerate != nil {
		request.FeeRate = feerate.String()
	}
	var result interface{}
	err := l.client.Request(request, &result)
	return result, err
}

type NewAddrRequest struct {
	AddressType string	`json:"addresstype,omitempty"`
}

func (r *NewAddrRequest) Name() string {
	return "newaddr"
}

type AddressType int

const (
	Bech32 AddressType = iota
	P2SHSegwit
)

func (a AddressType) String() string {
	return []string{"bech32", "p2sh-segwit"}[a]
}

// Get new Bech32 address for the internal wallet.
func (l *Lightning) NewAddr() (interface{}, error) {
	return l.NewAddressOfType(Bech32)
}

// Get new address of type {addrType} of the internal wallet.
func (l *Lightning) NewAddressOfType(addrType AddressType) (interface{}, error) {
	var result interface{}
	err := l.client.Request(&NewAddrRequest{addrType.String()}, result)
	return result, err
}

type ListFundsRequest struct {}

func (r *ListFundsRequest) Name() string {
	return "listfunds"
}

// Show funds available for opening channels
func (l *Lightning) ListFunds() (interface{}, error) {
	var result interface{}
	err := l.client.Request(&ListFundsRequest{}, result)
	return result, err
}

type ListForwardsRequest struct {}

func (r *ListForwardsRequest) Name() string {
	return "listforwards"
}

// List all forwarded payments and their information
func (l *Lightning) ListForwards() (interface{}, error) {
	var result interface{}
	err := l.client.Request(&ListForwardsRequest{}, result)
	return result, err
}

type DevRescanOutputsRequest struct {}

func (r *DevRescanOutputsRequest) Name() string {
	return "dev-rescan-outputs"
}

// Synchronize the state of our funds with bitcoind
func (l *Lightning) DevRescanOutputs() (interface{}, error) {
	var result interface{}
	err := l.client.Request(&DevRescanOutputsRequest{}, result)
	return result, err
}

type DevForgetChannelRequest struct {
	PeerId string	`json:"id"`
	Force bool	`json:"force"`
}

func (r *DevForgetChannelRequest) Name() string {
	return "dev-forget-channel"
}

// Forget channel with id {peerId}. Optionally {force} if has active channel.
// Caution, this might lose you funds.
func (l *Lightning) DevForgetChannel(peerId string, force bool) (interface{}, error) {
	var result interface{}
	err := l.client.Request(&DevForgetChannelRequest{peerId, force}, result)
	return result, err
}

type DisconnectRequest struct {
	PeerId string	`json:"id"`
	Force bool	`json:"force"`
}

func (r *DisconnectRequest) Name() string {
	return "disconnect"
}

// Disconnect from peer with {peerId}. Optionally {force} if has active channel.
func (l *Lightning) Disconnect(peerId string, force bool) (interface{}, error) {
	var result interface{}
	err := l.client.Request(&DisconnectRequest{peerId, force}, result)
	return result, err
}

type FeeRatesRequest struct {
	Style string	`json:"style"`
}

func (r *FeeRatesRequest) Name() string {
	return "feerates"
}

// Return feerate estimates, either satoshi-per-kw or satoshi-per-kb {style}
func (l *Lightning) FeeRates(style FeeRateStyle) (interface{}, error) {
	var result interface{}
	err := l.client.Request(&FeeRatesRequest{style.String()}, result)
	return result, err
}
