package pkg

import (
	"context"
	"errors"
	"fmt"

	envoycorev2 "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoyauthv2 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2"
	"github.com/solo-io/ext-auth-plugins/api"
	"github.com/solo-io/go-utils/contextutils"
	"go.uber.org/zap"
)

var (
	UnexpectedConfigError = func(typ interface{}) error {
		return errors.New(fmt.Sprintf("unexpected config type %T", typ))
	}
	_ api.ExtAuthPlugin = new(RequiredHeaderPlugin)
)

type RequiredHeaderPlugin struct{}

type Config struct {
	RequiredHeader string
	AllowedValues  []string
}

func (p *RequiredHeaderPlugin) NewConfigInstance(ctx context.Context) (interface{}, error) {
	return &Config{}, nil
}

func (p *RequiredHeaderPlugin) GetAuthService(ctx context.Context, configInstance interface{}) (api.AuthService, error) {
	config, ok := configInstance.(*Config)
	if !ok {
		return nil, UnexpectedConfigError(configInstance)
	}

	logger(ctx).Infow("Parsed RequiredHeaderAuthService config",
		zap.Any("requiredHeader", config.RequiredHeader),
		zap.Any("allowedHeaderValues", config.AllowedValues),
	)

	valueMap := map[string]bool{}
	for _, v := range config.AllowedValues {
		valueMap[v] = true
	}

	return &RequiredHeaderAuthService{
		RequiredHeader: config.RequiredHeader,
		AllowedValues:  valueMap,
	}, nil
}

type RequiredHeaderAuthService struct {
	RequiredHeader string
	AllowedValues  map[string]bool
}

// You can use the provided context to perform operations that are bound to the services lifecycle.
func (c *RequiredHeaderAuthService) Start(context.Context) error {
	// no-op
	return nil
}

func (c *RequiredHeaderAuthService) Authorize(ctx context.Context, request *api.AuthorizationRequest) (*api.AuthorizationResponse, error) {
	for key, value := range request.CheckRequest.GetAttributes().GetRequest().GetHttp().GetHeaders() {
		if key == c.RequiredHeader {
			logger(ctx).Infow("Found required header, checking value.", "header", key, "value", value)

			if _, ok := c.AllowedValues[value]; ok {
				logger(ctx).Infow("Header value match. Allowing request.")
				response := api.AuthorizedResponse()

				// Append extra header
				response.CheckResponse.HttpResponse = &envoyauthv2.CheckResponse_OkResponse{
					OkResponse: &envoyauthv2.OkHttpResponse{
						Headers: []*envoycorev2.HeaderValueOption{{
							Header: &envoycorev2.HeaderValue{
								Key:   "matched-allowed-headers",
								Value: "true",
							},
						}},
					},
				}
				return response, nil
			}
			logger(ctx).Infow("Header value does not match allowed values, denying access.")
			return api.UnauthorizedResponse(), nil
		}
	}
	logger(ctx).Infow("Required header not found, denying access")
	return api.UnauthorizedResponse(), nil
}

func logger(ctx context.Context) *zap.SugaredLogger {
	return contextutils.LoggerFrom(contextutils.WithLogger(ctx, "header_value_plugin"))
}
