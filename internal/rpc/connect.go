package rpc

import (
	"fmt"

	"github.com/doublemo/baa/internal/conf"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// NewConnect 创建连接
func NewConnect(c conf.RPCClient) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(makePolicy(c)), // This sets the initial balancing policy.
	}

	if len(c.Key) > 0 && len(c.Salt) > 0 {
		creds, err := credentials.NewClientTLSFromFile(c.Salt, c.Key)
		if err != nil {
			return nil, err
		}

		opts = append(opts, grpc.WithTransportCredentials(creds))
		opts = append(opts, grpc.WithPerRPCCredentials(oauth.NewOauthAccess(
			&oauth2.Token{AccessToken: c.ServiceSecurityKey},
		)))

	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	return grpc.Dial(fmt.Sprintf("%s:///%s", c.Name, c.Group), opts...)
}

func makePolicy(c conf.RPCClient) string {
	return `{
		"loadBalancingPolicy": "round_robin",
		"healthCheckConfig":{
			"serviceName" :""
		},
		"methodConfig": [{
			"name": [
				{"service": "sfu"},
				{"service": "auth"},
				{"service": "snid"},
				{"service": "im"},
				{"service": "usrt"},
				{"service": "imf"}
			],
			"waitForReady": true,
			"timeout":"1s",
			"maxRequestMessageBytes":10240,
			"maxResponseMessageBytes":10240,
	
			"retryPolicy": {
				"maxAttempts": 4,
				"initialBackoff": ".01s",
				"maxBackoff": ".01s",
				"backoffMultiplier": 1.0,
				"retryableStatusCodes": [ "UNAVAILABLE" ]
			},
	
			"hedgingPolicy":{
				"maxAttempts":4,
				"hedgingDelay": "0s",
				"nonFatalStatusCodes":  [ 
				"UNAVAILABLE",
				"INTERNAL",
				"ABORTED" ]
			}
		}],
	
		"retryThrottling":{
			"maxTokens": 10,
			"tokenRatio":0.1
		}
	}`
}
