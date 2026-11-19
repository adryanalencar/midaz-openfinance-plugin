package directory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestResolveBankSuccess tests the dynamic discovery of bank endpoints from the Open Finance Directory
func TestResolveBankSuccess(t *testing.T) {
	var directoryServer *httptest.Server
	directoryServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/organisations":
			participants := []DirectoryParticipant{
				{
					OrganisationID:     "org-itau-01",
					OrganisationName:   "ITAÚ UNIBANCO S.A.",
					RegistrationNumber: "60701190000104",
					Status:             "Active",
					AuthorisationServers: []AuthorisationServer{
						{
							AuthorisationServerID:   "as-itau-01",
							Name:                    "Itau Open Finance AS",
							OpenIDDiscoveryDocument: directoryServer.URL + "/.well-known/openid-configuration",
							ApiResources: []ApiResource{
								{
									ApiFamilyType: "accounts",
									ApiEndpoint:   "https://api.itau.com.br",
								},
							},
						},
					},
				},
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(participants)

		case "/.well-known/openid-configuration":
			config := WellKnownOpenIDConfig{
				Issuer:                "https://auth.itau.com.br",
				AuthorizationEndpoint: "https://auth.itau.com.br/auth",
				TokenEndpoint:         "https://auth.itau.com.br/token",
				JwksURI:               "https://auth.itau.com.br/jwks.json",
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(config)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer directoryServer.Close()

	client := NewDirectoryClient(directoryServer.Client(), directoryServer.URL)

	resolved, err := client.ResolveBank(context.Background(), "341") // 341 = Itaú

	assert.NoError(t, err)
	assert.NotNil(t, resolved)
	assert.Equal(t, "341", resolved.CompeCode)
	assert.Equal(t, "ITAÚ UNIBANCO S.A.", resolved.BrandName)
	assert.Equal(t, "https://api.itau.com.br", resolved.BaseApiURL)
	assert.Equal(t, "https://auth.itau.com.br/token", resolved.TokenURL)
	assert.Equal(t, "https://auth.itau.com.br/auth", resolved.AuthorisationURL)
	assert.Equal(t, "https://auth.itau.com.br/jwks.json", resolved.JwksURL)

	// Test cache hit (should return same instance without HTTP error even if server is closed)
	cached, err := client.ResolveBank(context.Background(), "341")
	assert.NoError(t, err)
	assert.Equal(t, resolved.TokenURL, cached.TokenURL)
}

// TestResolveBankNotFound tests error when bank is not present in directory
func TestResolveBankNotFound(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`)) // empty list
	}))
	defer mockServer.Close()

	client := NewDirectoryClient(mockServer.Client(), mockServer.URL)
	_, err := client.ResolveBank(context.Background(), "999")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found in Open Finance Directory")
}
