package resources

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestMapLabelMatchersFromNative(t *testing.T) {
	tests := []struct {
		name          string
		labelMatchers []string
		expected      LabelMatchers
		expectError   bool
	}{
		{
			name:          "ValidInput",
			labelMatchers: []string{"key1=value1,key1=value2", "key2=value3,key2=value4"},
			expected: LabelMatchers{
				{Key: types.StringValue("key1"), Values: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("value1"), types.StringValue("value2")})},
				{Key: types.StringValue("key2"), Values: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("value3"), types.StringValue("value4")})},
			},
			expectError: false,
		},
		{
			name:          "InvalidInput",
			labelMatchers: []string{"key1=value1,value2", "invalid"},
			expected:      nil,
			expectError:   true,
		},
		{
			name:          "EmptyInput",
			labelMatchers: []string{},
			expected:      LabelMatchers{},
			expectError:   false,
		},
		{
			name:          "EmptyValues",
			labelMatchers: []string{"key1="},
			expected: LabelMatchers{
				{Key: types.StringValue("key1"), Values: types.ListValueMust(types.StringType, []attr.Value{types.StringValue("")})},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, diags := MapLabelMatchersFromNative(ctx, tt.labelMatchers)

			if tt.expectError {
				assert.Nil(t, result)
				assert.True(t, diags.HasError())
			} else {
				assert.False(t, diags.HasError())
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
