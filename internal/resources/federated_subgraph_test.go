package resources

import (
	"github.com/labd/terraform-provider-wundergraph/sdk/wg/cosmo/common"
	platformv1 "github.com/labd/terraform-provider-wundergraph/sdk/wg/cosmo/platform/v1"
	"testing"

	_ "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestMapSubscriptionProtocol(t *testing.T) {
	tests := []struct {
		name        string
		protocol    types.String
		expected    common.GraphQLSubscriptionProtocol
		expectError bool
	}{
		{
			name:        "ValidInputWS",
			protocol:    types.StringValue("ws"),
			expected:    common.GraphQLSubscriptionProtocol_GRAPHQL_SUBSCRIPTION_PROTOCOL_WS,
			expectError: false,
		},
		{
			name:        "ValidInputSSE",
			protocol:    types.StringValue("sse"),
			expected:    common.GraphQLSubscriptionProtocol_GRAPHQL_SUBSCRIPTION_PROTOCOL_SSE,
			expectError: false,
		},
		{
			name:        "ValidInputSSEPost",
			protocol:    types.StringValue("sse_post"),
			expected:    common.GraphQLSubscriptionProtocol_GRAPHQL_SUBSCRIPTION_PROTOCOL_SSE_POST,
			expectError: false,
		},
		{
			name:        "InvalidInput",
			protocol:    types.StringValue("invalid"),
			expectError: true,
		},
		{
			name:        "NullInput",
			protocol:    types.StringNull(),
			expected:    common.GraphQLSubscriptionProtocol_GRAPHQL_SUBSCRIPTION_PROTOCOL_WS,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MapSubscriptionProtocol(tt.protocol)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, &tt.expected, result)
			}
		})
	}
}

func TestMapWebSocketSubprotocol(t *testing.T) {
	tests := []struct {
		name        string
		subprotocol types.String
		expected    common.GraphQLWebsocketSubprotocol
		expectError bool
	}{
		{
			name:        "ValidInputAuto",
			subprotocol: types.StringValue("auto"),
			expected:    common.GraphQLWebsocketSubprotocol_GRAPHQL_WEBSOCKET_SUBPROTOCOL_AUTO,
			expectError: false,
		},
		{
			name:        "ValidInputGraphQLWS",
			subprotocol: types.StringValue("graphql-ws"),
			expected:    common.GraphQLWebsocketSubprotocol_GRAPHQL_WEBSOCKET_SUBPROTOCOL_WS,
			expectError: false,
		},
		{
			name:        "ValidInputGraphQLTransportWS",
			subprotocol: types.StringValue("graphql-transport-ws"),
			expected:    common.GraphQLWebsocketSubprotocol_GRAPHQL_WEBSOCKET_SUBPROTOCOL_TRANSPORT_WS,
			expectError: false,
		},
		{
			name:        "InvalidInput",
			subprotocol: types.StringValue("invalid"),
			expectError: true,
		},
		{
			name:        "NullInput",
			subprotocol: types.StringNull(),
			expected:    common.GraphQLWebsocketSubprotocol_GRAPHQL_WEBSOCKET_SUBPROTOCOL_AUTO,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MapWebSocketSubprotocol(tt.subprotocol)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, &tt.expected, result)
			}
		})
	}
}

func TestMapLabelsToNative(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		expected []*platformv1.Label
	}{
		{
			name:     "ValidInput",
			labels:   map[string]string{"key1": "value1", "key2": "value2"},
			expected: []*platformv1.Label{{Key: "key1", Value: "value1"}, {Key: "key2", Value: "value2"}},
		},
		{
			name:   "EmptyInput",
			labels: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapLabelsToNative(tt.labels)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapLabelsFromNative(t *testing.T) {
	tests := []struct {
		name     string
		labels   []*platformv1.Label
		expected map[string]string
	}{
		{
			name:     "ValidInput",
			labels:   []*platformv1.Label{{Key: "key1", Value: "value1"}, {Key: "key2", Value: "value2"}},
			expected: map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name:     "EmptyInput",
			labels:   []*platformv1.Label{},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapLabelsFromNative(tt.labels)
			assert.Equal(t, tt.expected, result)
		})
	}
}
