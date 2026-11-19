package directory

// WellKnownOpenIDConfig holds standard OIDC discovery endpoints returned by the bank
type WellKnownOpenIDConfig struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	UserinfoEndpoint      string   `json:"userinfo_endpoint"`
	JwksURI               string   `json:"jwks_uri"`
	RegistrationEndpoint  string   `json:"registration_endpoint"`
	ScopesSupported       []string `json:"scopes_supported"`
	ResponseTypesSupported []string `json:"response_types_supported"`
}

// DirectoryParticipant represents an enrolled financial institution in the Open Finance Brasil Directory
type DirectoryParticipant struct {
	OrganisationID        string                `json:"OrganisationId"`
	OrganisationName      string                `json:"OrganisationName"`
	RegistrationNumber    string                `json:"RegistrationNumber"` // CNPJ
	Status                string                `json:"Status"`             // Active
	AuthorisationServers  []AuthorisationServer `json:"AuthorisationServers"`
}

type AuthorisationServer struct {
	AuthorisationServerID  string        `json:"AuthorisationServerId"`
	Name                   string        `json:"Name"`
	OpenIDDiscoveryDocument string        `json:"OpenIDDiscoveryDocument"`
	ApiResources           []ApiResource `json:"ApiResources"`
}

type ApiResource struct {
	ApiFamilyType string `json:"ApiFamilyType"` // "accounts", "consents", "payments-pix"
	ApiVersion    string `json:"ApiVersion"`
	ApiEndpoint   string `json:"ApiEndpoint"`
}

// ResolvedBankEndpoints holds the discovered URLs for a target bank
type ResolvedBankEndpoints struct {
	CompeCode             string
	BrandName             string
	BaseApiURL            string
	TokenURL              string
	AuthorisationURL      string
	JwksURL               string
}
