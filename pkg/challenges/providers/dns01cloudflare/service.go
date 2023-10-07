package dns01cloudflare

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/httpclient"
	"legocerthub-backend/pkg/output"

	"github.com/cloudflare/cloudflare-go"
	"go.uber.org/zap"
)

var (
	errServiceComponent = errors.New("necessary dns-01 cloudflare challenge service component is missing")
)

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetHttpClient() *httpclient.Client
}

// provider Service struct
type Service struct {
	logger        *zap.SugaredLogger
	httpClient    *httpclient.Client
	cloudflareApi *cloudflare.API
	domainIDs     map[string]string // domain_name[zone_id]
}

// ChallengeType returns the ACME Challenge Type this provider uses, which is dns-01
func (service *Service) AcmeChallengeType() acme.ChallengeType {
	return acme.ChallengeTypeDns01
}

// Stop is used for any actions needed prior to deleting this provider. If no actions
// are needed, it is just a no-op.
func (service *Service) Stop() error { return nil }

// NewService creates a new instance of the Cloudflare provider service. Service
// contains one Cloudflare API instance.
func NewService(app App, cfg *Config) (*Service, error) {
	// if no config, error
	if cfg == nil {
		return nil, errServiceComponent
	}

	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// http client for api calls
	service.httpClient = app.GetHttpClient()

	// make map for domains
	service.domainIDs = make(map[string]string)

	// cloudflare api
	err := service.configureCloudflareAPI(cfg)
	if err != nil {
		return nil, err
	}

	// debug log configured domains
	service.logger.Infof("cloudflare instance %s configured domains: %s", service.redactedApiIdentifier(), cfg.Doms)

	return service, nil
}

// Update Service updates the Service to use the new config
func (service *Service) UpdateService(app App, cfg *Config) error {
	// if form submitted with redacted info, ensure unredacted is submitted to new service
	if cfg.Account != nil && cfg.Account.GlobalApiKey.Redacted() == service.redactedApiIdentifier() {
		*cfg.Account.GlobalApiKey = output.RedactedString(service.apiIdentifier())

	} else if cfg.ApiToken != nil && cfg.ApiToken.Redacted() == service.redactedApiIdentifier() {
		*cfg.ApiToken = output.RedactedString(service.apiIdentifier())
	}

	// don't need to do anything with "old" Service, just set a new one
	newServ, err := NewService(app, cfg)
	if err != nil {
		return err
	}

	// set content of old pointer so anything with the pointer calls the
	// updated service
	*service = *newServ

	return nil
}

// apiIdentifier selects either the APIKey, APIUserServiceKey, or APIToken
// (depending on which is in use for the API instance) and returns it.
func (service *Service) apiIdentifier() string {
	// return whichever is present
	if len(service.cloudflareApi.APIToken) > 0 {
		return service.cloudflareApi.APIToken
	} else if len(service.cloudflareApi.APIKey) > 0 {
		return service.cloudflareApi.APIKey
	} else if len(service.cloudflareApi.APIUserServiceKey) > 0 {
		return service.cloudflareApi.APIUserServiceKey
	}

	// none present, return unknown
	return "unknown"
}

// redactedApiIdentifier selects either the APIKey, APIUserServiceKey, or APIToken
// (depending on which is in use for the API instance) and then redacts it to return
// the first and last characters of the key separated with asterisks. This is useful
// for logging issues without saving the full credential to logs.
func (service *Service) redactedApiIdentifier() string {
	return output.RedactString(service.apiIdentifier())
}
