package resources

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/terraform-provider-wundergraph/sdk/wg/cosmo/common"
	platformv1 "github.com/labd/terraform-provider-wundergraph/sdk/wg/cosmo/platform/v1"
	"github.com/labd/terraform-provider-wundergraph/sdk/wg/cosmo/platform/v1/platformv1connect"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &FederatedSubgraphResource{}
var _ resource.ResourceWithImportState = &FederatedSubgraphResource{}

func NewFederatedSubgraphResource() resource.Resource {
	return &FederatedSubgraphResource{}
}

// FederatedSubgraphResource defines the resource implementation.
type FederatedSubgraphResource struct {
	client platformv1connect.PlatformServiceClient
}

// FederatedSubgraphModel describes the resource data model.
type FederatedSubgraphModel struct {
	Id                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Namespace            types.String `tfsdk:"namespace"`
	RoutingUrl           types.String `tfsdk:"routing_url"`
	Schema               types.String `tfsdk:"schema"`
	SubscriptionUrl      types.String `tfsdk:"subscription_url"`
	SubscriptionProtocol types.String `tfsdk:"subscription_protocol"`
	WebsocketSubprotocol types.String `tfsdk:"websocket_subprotocol"`
	Labels               types.Map    `tfsdk:"labels"`
	IsEventDrivenGraph   types.Bool   `tfsdk:"is_event_driven_graph"`
	IsFeatureSubgraph    types.Bool   `tfsdk:"is_feature_subgraph"`
}

func (r *FederatedSubgraphResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_federated_subgraph"
}

func (r *FederatedSubgraphResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "federated subgraph resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the subgraph to create. It is usually in the format of <org>.<service.name> and is used to uniquely identify your federated subgraph.",
				Required:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The namespace name of the subgraph. Defaults to default.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
			},
			"routing_url": schema.StringAttribute{
				MarkdownDescription: "The routing URL of your subgraph. This is the url at which the subgraph will be accessible. Required unless the event-driven-graph flag is set. Returns an error if the event-driven-graph flag is set.",
				Optional:            true,
			},
			"schema": schema.StringAttribute{
				MarkdownDescription: "The schema to upload to the subgraph. This should be the full schema in SDL format.",
				Required:            true,
			},
			"subscription_url": schema.StringAttribute{
				MarkdownDescription: "The protocol to use when subscribing to the subgraph. The supported protocols are ws, sse, and sse_post. Returns an error if the event-driven-graph flag is set.",
				Optional:            true,
			},
			"subscription_protocol": schema.StringAttribute{
				MarkdownDescription: "The protocol to use when subscribing to the subgraph. The supported protocols are ws, sse, and sse_post.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("ws"),
				Validators: []validator.String{
					stringvalidator.OneOf("ws", "sse", "sse_post"),
				},
			},
			"websocket_subprotocol": schema.StringAttribute{
				MarkdownDescription: "The subprotocol to use when subscribing to the subgraph. The supported protocols are auto, graphql-ws, and graphql-transport-ws. Should be used only if the subscription protocol is ws. For more information see https://cosmo-docs.wundergraph.com/router/subscriptions/websocket-subprotocols.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("auto"),
				Validators: []validator.String{
					stringvalidator.OneOf("auto", "graphql-ws", "graphql-transport-ws"),
				},
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The labels to apply to the subgraph.",
				Optional:            true,
			},
			//Readme cannot be implemented because the get endpoint does not return it, so we don't know the actual state
			//"readme": schema.StringAttribute{
			//	MarkdownDescription: "The markdown text which describes the subgraph",
			//	Optional:            true,
			//},
			"is_event_driven_graph": schema.BoolAttribute{
				MarkdownDescription: "Set whether the subgraph is an Event-Driven Graph (EDG). Errors will be returned for the inclusion of most other parameters if the subgraph is an Event-Driven Graph.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"is_feature_subgraph": schema.BoolAttribute{
				MarkdownDescription: "Set whether the subgraph is a feature subgraph.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *FederatedSubgraphResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(platformv1connect.PlatformServiceClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func MapSubscriptionProtocol(sp types.String) (*common.GraphQLSubscriptionProtocol, error) {
	if sp.IsNull() || sp.IsUnknown() {
		p := common.GraphQLSubscriptionProtocol_GRAPHQL_SUBSCRIPTION_PROTOCOL_WS
		return &p, nil
	}

	var p common.GraphQLSubscriptionProtocol
	switch sp.ValueString() {
	case "ws":
		p = common.GraphQLSubscriptionProtocol_GRAPHQL_SUBSCRIPTION_PROTOCOL_WS
	case "sse":
		p = common.GraphQLSubscriptionProtocol_GRAPHQL_SUBSCRIPTION_PROTOCOL_SSE
	case "sse_post":
		p = common.GraphQLSubscriptionProtocol_GRAPHQL_SUBSCRIPTION_PROTOCOL_SSE_POST
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", sp.ValueString())
	}

	return &p, nil
}

func MapWebSocketSubprotocol(wsp types.String) (*common.GraphQLWebsocketSubprotocol, error) {
	if wsp.IsNull() || wsp.IsUnknown() {
		var p = common.GraphQLWebsocketSubprotocol_GRAPHQL_WEBSOCKET_SUBPROTOCOL_AUTO
		return &p, nil
	}

	var p common.GraphQLWebsocketSubprotocol
	switch wsp.ValueString() {
	case "auto":
		p = common.GraphQLWebsocketSubprotocol_GRAPHQL_WEBSOCKET_SUBPROTOCOL_AUTO
	case "graphql-ws":
		p = common.GraphQLWebsocketSubprotocol_GRAPHQL_WEBSOCKET_SUBPROTOCOL_WS
	case "graphql-transport-ws":
		p = common.GraphQLWebsocketSubprotocol_GRAPHQL_WEBSOCKET_SUBPROTOCOL_TRANSPORT_WS
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", wsp.ValueString())
	}

	return &p, nil
}

func MapLabelsToNative(labels map[string]string) []*platformv1.Label {
	var l []*platformv1.Label
	for k, v := range labels {
		l = append(l, &platformv1.Label{
			Key:   k,
			Value: v,
		})
	}

	return l
}

func (r *FederatedSubgraphResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *FederatedSubgraphModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	p, err := MapSubscriptionProtocol(plan.SubscriptionProtocol)
	if err != nil {
		resp.Diagnostics.AddError("Error creating subgraph", err.Error())
		return
	}

	w, err := MapWebSocketSubprotocol(plan.WebsocketSubprotocol)
	if err != nil {
		resp.Diagnostics.AddError("Error creating subgraph", err.Error())
		return
	}

	var labels map[string]string
	diags = plan.Labels.ElementsAs(ctx, &labels, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	//We first create the subgraph
	rc, err := r.client.CreateFederatedSubgraph(ctx, &connect.Request[platformv1.CreateFederatedSubgraphRequest]{
		Msg: &platformv1.CreateFederatedSubgraphRequest{
			Name:                 plan.Name.ValueString(),
			Namespace:            plan.Namespace.ValueString(),
			RoutingUrl:           plan.RoutingUrl.ValueStringPointer(),
			SubscriptionUrl:      plan.SubscriptionUrl.ValueStringPointer(),
			SubscriptionProtocol: p,
			WebsocketSubprotocol: w,
			Labels:               MapLabelsToNative(labels),
			IsEventDrivenGraph:   plan.IsEventDrivenGraph.ValueBoolPointer(),
			IsFeatureSubgraph:    plan.IsFeatureSubgraph.ValueBoolPointer(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating subgraph", err.Error())
		return
	}

	if rc.Msg.GetResponse().Code != common.EnumStatusCode_OK {
		resp.Diagnostics.AddError("Error creating subgraph", rc.Msg.GetResponse().GetDetails())
		return
	}

	// We fetch the subgraph list to get the requested subgraph, as we don't have a direct read endpoint.
	ns, err := r.client.GetSubgraphs(ctx, &connect.Request[platformv1.GetSubgraphsRequest]{
		Msg: &platformv1.GetSubgraphsRequest{
			Namespace: plan.Namespace.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading subgraph", err.Error())
		return
	}

	var isFound bool
	for _, n := range ns.Msg.Graphs {
		if n.Name == plan.Name.ValueString() {
			plan.Id = types.StringValue(n.Id)
			isFound = true
			continue
		}
	}

	if !isFound {
		resp.Diagnostics.AddError("Error reading subgraph", "subgraph not found")
		return
	}

	// We now publish the first subgraph
	rp, err := r.client.PublishFederatedSubgraph(ctx, &connect.Request[platformv1.PublishFederatedSubgraphRequest]{
		Msg: &platformv1.PublishFederatedSubgraphRequest{
			Name:      plan.Name.ValueString(),
			Namespace: plan.Namespace.ValueString(),
			Schema:    plan.Schema.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating schema", err.Error())
		return
	}

	if rp.Msg.GetResponse().Code != common.EnumStatusCode_OK {
		resp.Diagnostics.AddError("Error creating subgraph", rp.Msg.GetResponse().GetDetails())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func MapLabelsFromNative(labels []*platformv1.Label) map[string]string {
	l := make(map[string]string)
	for _, v := range labels {
		l[v.Key] = v.Value
	}

	return l
}

func (r *FederatedSubgraphResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *FederatedSubgraphModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// We fetch the subgraph list to get the requested subgraph, as we don't have a direct read endpoint.
	ns, err := r.client.GetSubgraphs(ctx, &connect.Request[platformv1.GetSubgraphsRequest]{
		Msg: &platformv1.GetSubgraphsRequest{
			Namespace: data.Namespace.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading subgraph", err.Error())
		return
	}

	if ns.Msg.GetResponse().Code != common.EnumStatusCode_OK {
		resp.Diagnostics.AddError("Error fetching subgraph list", ns.Msg.GetResponse().GetDetails())
		return
	}

	var current *FederatedSubgraphModel
	for _, n := range ns.Msg.Graphs {
		if n.Id == data.Id.ValueString() {
			// We need to check if the subscription url is empty, as it is optional. If an empty string is returned we assume it is nil.
			var subscriptionUrl *string = nil
			if n.SubscriptionUrl != "" {
				subscriptionUrl = &n.SubscriptionUrl
			}

			labels, diags := types.MapValueFrom(ctx, types.StringType, MapLabelsFromNative(n.Labels))
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}

			current = &FederatedSubgraphModel{
				Id:                   types.StringValue(n.Id),
				Name:                 types.StringValue(n.Name),
				Namespace:            types.StringValue(n.Namespace),
				RoutingUrl:           types.StringValue(n.RoutingURL),
				SubscriptionUrl:      types.StringPointerValue(subscriptionUrl),
				IsEventDrivenGraph:   types.BoolValue(n.IsEventDrivenGraph),
				IsFeatureSubgraph:    types.BoolValue(n.IsFeatureSubgraph),
				SubscriptionProtocol: types.StringValue(n.SubscriptionProtocol),
				WebsocketSubprotocol: types.StringValue(n.WebsocketSubprotocol),
				Labels:               labels,
			}
			continue
		}
	}

	if current == nil {
		resp.Diagnostics.AddError("Error reading federated subgraph", "federated subgraph not found")
		return
	}

	sdl, err := r.client.GetLatestSubgraphSDL(ctx, &connect.Request[platformv1.GetLatestSubgraphSDLRequest]{
		Msg: &platformv1.GetLatestSubgraphSDLRequest{
			Namespace: data.Namespace.ValueString(),
			Name:      data.Name.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error fetching sdl", err.Error())
		return
	}

	if sdl.Msg.GetResponse().Code != common.EnumStatusCode_OK {
		resp.Diagnostics.AddError("Error fetching SDL", sdl.Msg.GetResponse().GetDetails())
		return
	}

	current.Schema = types.StringPointerValue(sdl.Msg.Sdl)

	resp.Diagnostics.Append(resp.State.Set(ctx, &current)...)
}

func (r *FederatedSubgraphResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FederatedSubgraphModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state FederatedSubgraphModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//TODO: check if a rename is intended. If that is the case we delete the old subgraph and create a new one.

	//First we update the schema itself
	if !plan.Schema.Equal(state.Schema) {
		rp, err := r.client.PublishFederatedSubgraph(ctx, &connect.Request[platformv1.PublishFederatedSubgraphRequest]{
			Msg: &platformv1.PublishFederatedSubgraphRequest{
				Name:      plan.Name.ValueString(),
				Namespace: plan.Namespace.ValueString(),
				Schema:    plan.Schema.ValueString(),
			},
		})
		if err != nil {
			resp.Diagnostics.AddError("Error updating schema", err.Error())
			return
		}

		if rp.Msg.GetResponse().Code != common.EnumStatusCode_OK {
			resp.Diagnostics.AddError("Error updating subgraph", rp.Msg.GetResponse().GetDetails())
			return
		}
	}

	//Then we check if the namespace needs to be moved
	if !plan.Namespace.Equal(state.Namespace) {
		rm, err := r.client.MoveSubgraph(ctx, &connect.Request[platformv1.MoveGraphRequest]{
			Msg: &platformv1.MoveGraphRequest{
				Name:         state.Name.ValueString(),
				Namespace:    state.Namespace.ValueString(),
				NewNamespace: plan.Namespace.ValueString(),
			},
		})
		if err != nil {
			resp.Diagnostics.AddError("Error moving namespace", err.Error())
			return
		}

		if rm.Msg.GetResponse().Code != common.EnumStatusCode_OK {
			resp.Diagnostics.AddError("Error updating subgraph", rm.Msg.GetResponse().GetDetails())
			return
		}
	}

	p, err := MapSubscriptionProtocol(plan.SubscriptionProtocol)
	if err != nil {
		resp.Diagnostics.AddError("Error creating subgraph", err.Error())
		return
	}

	w, err := MapWebSocketSubprotocol(plan.WebsocketSubprotocol)
	if err != nil {
		resp.Diagnostics.AddError("Error creating subgraph", err.Error())
		return
	}

	var labels map[string]string
	diags := plan.Labels.ElementsAs(ctx, &labels, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	ru, err := r.client.UpdateSubgraph(ctx, &connect.Request[platformv1.UpdateSubgraphRequest]{
		Msg: &platformv1.UpdateSubgraphRequest{
			Name:                 plan.Name.ValueString(),
			Namespace:            plan.Namespace.ValueString(),
			RoutingUrl:           plan.RoutingUrl.ValueStringPointer(),
			SubscriptionUrl:      plan.SubscriptionUrl.ValueStringPointer(),
			Labels:               MapLabelsToNative(labels),
			SubscriptionProtocol: p,
			WebsocketSubprotocol: w,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading subgraph", err.Error())
		return
	}

	if ru.Msg.GetResponse().Code != common.EnumStatusCode_OK {
		resp.Diagnostics.AddError("Error updating subgraph", ru.Msg.GetResponse().GetDetails())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FederatedSubgraphResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *FederatedSubgraphModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteFederatedSubgraph(ctx, &connect.Request[platformv1.DeleteFederatedSubgraphRequest]{
		Msg: &platformv1.DeleteFederatedSubgraphRequest{
			SubgraphName: data.Name.ValueString(),
			Namespace:    data.Namespace.ValueString(),
		},
	})

	if err != nil {
		resp.Diagnostics.AddError("Error deleting subgraph", err.Error())
		return
	}
}

func (r *FederatedSubgraphResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
