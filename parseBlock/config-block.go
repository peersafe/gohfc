package parseBlock

type Config struct {
	ChannelGroup *ConfigGroup `json:"channel_group,omitempty"`
	Type         int32        `json:"type,omitempty"`
	Sequence     string       `json:"sequence,omitempty"`
}

type ConfigGroup struct {
	Groups    map[string]*ConfigGroup  `json:"groups,omitempty"`
	ModPolicy string                   `json:"mod_policy,omitempty"`
	Policies  map[string]*ConfigPolicy `json:"policies,omitempty"`
	Values    map[string]*Value        `json:"values,omitempty"`
	Version   string                   `json:"version,omitempty"`
}

type ConfigPolicy struct {
	ModPolicy string  `json:"mod_policy,omitempty"`
	Policy    *Policy `json:"policy,omitempty"`
	Version   string  `json:"version,omitempty"`
}

type Policy struct {
	Type      int32  `json:"type,omitempty"`
	ModPolicy string `json:"mod_policy,omitempty"`
	Value     *Value `json:"value,omitempty"`
}

type FabricNodeOUs struct {
	Enable             bool          `json:"Enable,omitempty"`
	ClientOUIdentifier *OUIdentifier `json:"clientOUIdentifier,omitempty"`
	PeerOUIdentifier   *OUIdentifier `json:"peerOUIdentifier,omitempty"`
}

type OUIdentifier struct {
	Certificate                  string `json:"certificate,omitempty"`
	OrganizationalUnitIdentifier string `json:"organizational_unit_identifier,omitempty"`
}

type MSP struct {
	FabricNodeOUs *FabricNodeOUs `json:"FabricNodeOUs,omitempty"`
	Admins        []string       `json:"admins,omitempty"`
	CryptoConfig  *CryptoConfig  `json:"crypto_config,omitempty"`
	Name          string         `json:"name,omitempty"`
	RootCerts     []string       `json:"root_certs,omitempty"`
	TlsRootCerts  []string       `json:"tls_root_certs,omitempty"`
}

type CryptoConfig struct {
	IdentityIdentifierHashFunction string `json:"identity_identifier_hash_function,omitempty"`
	SignatureHashFamily            string `json:"signature_hash_family,omitempty"`
}

type AnchorPeer struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type Value struct {
	ModPolicy         string        `json:"mod_policy,omitempty"`
	Identities        []*Identitie  `json:"identities,omitempty"`
	Capabilities      interface{}   `json:"capabilities,omitempty"`
	Rule              interface{}   `json:"rule,omitempty"`
	SubPolicy         string        `json:"sub_policy,omitempty"`
	Config            *MSP          `json:"config,omitempty"`
	AnchorPeers       []*AnchorPeer `json:"anchor_peers,omitempty"`
	MSP               *Value        `json:"MSP,omitempty"`
	AbsoluteMaxBytes  int64         `json:"absolute_max_bytes,omitempty"`
	MaxMessageCount   int64         `json:"max_message_count,omitempty"`
	PreferredMaxBytes int64         `json:"preferred_max_bytes,omitempty"`
	Timeout           string        `json:"timeout,omitempty"`
	Brokers           []string      `json:"brokers,omitempty"`
	Width             int64         `json:"width,omitempty"`
	Name              string        `json:"name,omitempty"`
	Addresses         []string      `json:"addresses,omitempty"`
	Value             *Value        `json:"value,omitempty"`
	Version           interface{}   `json:"version,omitempty"`
	Type              interface{}   `json:"type,omitempty"`
}

type Identitie struct {
	Principal               Principal `json:"principal,omitempty"`
	PrincipalClassification string    `json:"principal_classification,omitempty"`
}

type Principal struct {
	MspIdentifier string `json:"msp_identifier,omitempty"`
	Role          string `json:"role,omitempty"`
}

type Rule struct {
	SignedBy int    `json:"signed_by,omitempty"`
	NOutOf   NOutOf `json:"n_out_of,omitempty"`
}

type NOutOf struct {
	N     int     `json:"n,omitempty"`
	Rules []*Rule `json:"rules,omitempty"`
}
