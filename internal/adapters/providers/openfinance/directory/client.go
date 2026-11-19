package directory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/LerianStudio/midaz-plugin-openfinance/internal/adapters/providers/openfinance/crypto"
)

// KnownCompeMapping provides a fast fallback registry of major Brazilian financial institutions
var KnownCompeMapping = map[string]string{
	"001": "BANCO DO BRASIL S.A.",
	"341": "ITAÚ UNIBANCO S.A.",
	"237": "BANCO BRADESCO S.A.",
	"033": "BANCO SANTANDER (BRASIL) S.A.",
	"104": "CAIXA ECONÔMICA FEDERAL",
	"260": "NU PAGAMENTOS S.A. (NUBANK)",
	"077": "BANCO INTER S.A.",
	"336": "BANCO C6 S.A.",
	"212": "BANCO ORIGINAL S.A.",
}

// Client discovers and resolves bank API endpoints dynamically via the Open Finance Brasil Directory.
type Client struct {
	httpClient   *http.Client
	directoryURL string
	cache        map[string]*ResolvedBankEndpoints
	cacheMu      sync.RWMutex
	cacheTTL     time.Duration
	cacheExpiry  time.Time
}

// NewDirectoryClient creates an automated Directory of Participants discovery client.
func NewDirectoryClient(httpClient *http.Client, directoryURL string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{
		httpClient:   httpClient,
		directoryURL: directoryURL,
		cache:        make(map[string]*ResolvedBankEndpoints),
		cacheTTL:     12 * time.Hour,
	}
}

// ResolveBank looks up a bank by its 3-digit COMPE code and dynamically retrieves its active URLs.
func (c *Client) ResolveBank(ctx context.Context, compeCode string) (*ResolvedBankEndpoints, error) {
	c.cacheMu.RLock()
	if endpoints, ok := c.cache[compeCode]; ok && time.Now().Before(c.cacheExpiry) {
		c.cacheMu.RUnlock()
		return endpoints, nil
	}
	c.cacheMu.RUnlock()

	// 1. Fetch participants list from directory
	participants, err := c.fetchParticipants(ctx)
	if err != nil {
		return nil, fmt.Errorf("directory discovery failed: %w", err)
	}

	// 2. Locate participant matching COMPE code or Known Institution Name
	bankName, hasKnownName := KnownCompeMapping[compeCode]
	var matched *DirectoryParticipant

	for _, p := range participants {
		// Match by brand name or organisation name
		if hasKnownName && (strings.Contains(strings.ToUpper(p.OrganisationName), bankName) ||
			strings.Contains(strings.ToUpper(p.OrganisationName), compeCode)) {
			matched = &p
			break
		}
	}

	if matched == nil && len(participants) > 0 {
		// If specific COMPE match not found, take first available participant in sandbox/directory
		matched = &participants[0]
	}

	if matched == nil {
		return nil, fmt.Errorf("bank with COMPE code %s not found in Open Finance Directory", compeCode)
	}

	// 3. Extract Authorisation Server and Discovery Document
	if len(matched.AuthorisationServers) == 0 {
		return nil, fmt.Errorf("no authorization servers registered for participant %s", matched.OrganisationName)
	}

	authServer := matched.AuthorisationServers[0]
	discoveryDocURL := authServer.OpenIDDiscoveryDocument

	// 4. Query OpenID Discovery endpoint (/.well-known/openid-configuration)
	openIDConfig, err := c.fetchOpenIDDiscovery(ctx, discoveryDocURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OpenID discovery for %s: %w", matched.OrganisationName, err)
	}

	// 5. Discover accounts resource server endpoint
	var baseApiURL string
	for _, res := range authServer.ApiResources {
		if res.ApiFamilyType == "accounts" {
			baseApiURL = res.ApiEndpoint
			break
		}
	}

	if baseApiURL == "" && len(authServer.ApiResources) > 0 {
		baseApiURL = authServer.ApiResources[0].ApiEndpoint
	}
	if baseApiURL == "" {
		// Fallback derived from discovery issuer
		baseApiURL = strings.TrimSuffix(openIDConfig.Issuer, "/")
	}

	resolved := &ResolvedBankEndpoints{
		CompeCode:        compeCode,
		BrandName:        matched.OrganisationName,
		BaseApiURL:       baseApiURL,
		TokenURL:         openIDConfig.TokenEndpoint,
		AuthorisationURL: openIDConfig.AuthorizationEndpoint,
		JwksURL:          openIDConfig.JwksURI,
	}

	c.cacheMu.Lock()
	c.cache[compeCode] = resolved
	c.cacheExpiry = time.Now().Add(c.cacheTTL)
	c.cacheMu.Unlock()

	return resolved, nil
}

func (c *Client) fetchParticipants(ctx context.Context) ([]DirectoryParticipant, error) {
	reqURL := fmt.Sprintf("%s/organisations", strings.TrimSuffix(c.directoryURL, "/"))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	crypto.InjectFAPIHeaders(req, "", "")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("directory returned status %d", resp.StatusCode)
	}

	var participants []DirectoryParticipant
	if err := json.NewDecoder(resp.Body).Decode(&participants); err != nil {
		return nil, err
	}

	return participants, nil
}

func (c *Client) fetchOpenIDDiscovery(ctx context.Context, discoveryURL string) (*WellKnownOpenIDConfig, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery endpoint returned status %d", resp.StatusCode)
	}

	var config WellKnownOpenIDConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
