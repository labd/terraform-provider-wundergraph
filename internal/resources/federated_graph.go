package resources

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/terraform-provider-wundergraph/sdk/wg/cosmo/common"
	platformv1 "github.com/labd/terraform-provider-wundergraph/sdk/wg/cosmo/platform/v1"
	"github.com/labd/terraform-provider-wundergraph/sdk/wg/cosmo/platform/v1/platformv1connect"
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &FederatedGraphResource{}
var _ resource.ResourceWithImportState = &FederatedGraphResource{}

func NewFederatedGraphResource() resource.Resource {
	return &FederatedGraphResource{}
}

// FederatedGraphResource defines the resource implementation.
type FederatedGraphResource struct {
	client platformv1connect.PlatformServiceClient
}

// FederatedGraphModel describes the resource data model.
type FederatedGraphModel struct {
	Id                     types.String  `tfsdk:"id"`
	Name                   types.String  `tfsdk:"name"`
	Namespace              types.String  `tfsdk:"namespace"`
	RoutingUrl             types.String  `tfsdk:"routing_url"`
	LabelMatchers          LabelMatchers `tfsdk:"label_matchers"`
	AdmissionWebhookUrl    types.String  `tfsdk:"admission_webhook_url"`
	AdmissionWebhookSecret types.String  `tfsdk:"admission_webhook_secret"`
}

type LabelMatcher struct {
	Key    types.String `tfsdk:"key"`
	Values types.List   `tfsdk:"values"`
}

type LabelMatchers []LabelMatcher

func (l *LabelMatchers) FindByKey(key string) *LabelMatcher {
	for _, v := range *l {
		if v.Key.ValueString() == key {
			return &v
		}
	}

	return nil
}

func (r *FederatedGraphResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_federated_graph"
}

func (r *FederatedGraphResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Federated graph.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the federated graph to create. It is usually in the format of <org>.<env> and is used to uniquely identify your federated graph.",
				Required:            true,
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The namespace name of the federated graph. Defaults to `default`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("default"),
			},
			"routing_url": schema.StringAttribute{
				MarkdownDescription: "The routing url of your router. This is the url that the router will be accessible at.",
				Required:            true,
			},
			"label_matchers": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The key of the label matcher.",
						},
						"values": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "The key of the label matcher.",
							Required:            true,
						},
					},
				},
				MarkdownDescription: "The label matcher is used to select the subgraphs to federate.",
				Optional:            true,
			},
			"admission_webhook_url": schema.StringAttribute{
				MarkdownDescription: "The admission webhook url. This is the url that the controlplane will use to implement admission control for the federated graph.",
				Optional:            true,
			},
			"admission_webhook_secret": schema.StringAttribute{
				MarkdownDescription: "The admission webhook secret is used to sign requests to the webhook url.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *FederatedGraphResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func MapLabelMatchersToNative(ctx context.Context, labelMatchers []LabelMatcher) ([]string, diag.Diagnostics) {
	var matchers []string
	for _, v := range labelMatchers {
		var values []string
		diags := v.Values.ElementsAs(ctx, &values, false)
		if diags.HasError() {
			return nil, diags
		}

		var tags []string
		for _, tag := range values {
			tags = append(tags, fmt.Sprintf("%s=%s", v.Key.ValueString(), tag))
		}

		matchers = append(matchers, strings.Join(tags, ","))
	}

	return matchers, nil
}

func (r *FederatedGraphResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *FederatedGraphModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)

	var labels, d = MapLabelMatchersToNative(ctx, plan.LabelMatchers)
	if diags.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	//We first create the FederatedGraph
	rc, err := r.client.CreateFederatedGraph(ctx, &connect.Request[platformv1.CreateFederatedGraphRequest]{
		Msg: &platformv1.CreateFederatedGraphRequest{
			Name:                   plan.Name.ValueString(),
			Namespace:              plan.Namespace.ValueString(),
			RoutingUrl:             plan.RoutingUrl.ValueString(),
			AdmissionWebhookURL:    plan.AdmissionWebhookUrl.ValueString(),
			AdmissionWebhookSecret: plan.AdmissionWebhookSecret.ValueStringPointer(),
			LabelMatchers:          labels,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating federated graph", err.Error())
		return
	}

	if rc.Msg.GetResponse().Code != common.EnumStatusCode_OK {
		for _, e := range rc.Msg.CompositionErrors {
			resp.Diagnostics.AddWarning("Composition errors when creating graph", e.Message)
		}

		for _, e := range rc.Msg.DeploymentErrors {
			resp.Diagnostics.AddWarning("Deployment errors when creating graph", e.Message)
		}

		resp.Diagnostics.AddError("Error creating federated graph", rc.Msg.GetResponse().GetDetails())
	}

	// We fetch the FederatedGraph list to get the requested FederatedGraph, as we don't have a direct read endpoint.
	ns, err := r.client.GetFederatedGraphs(ctx, &connect.Request[platformv1.GetFederatedGraphsRequest]{
		Msg: &platformv1.GetFederatedGraphsRequest{
			Namespace: plan.Namespace.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading federated graph", err.Error())
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
		resp.Diagnostics.AddError("Error reading federated graph", "Federated graph not found")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func MapLabelMatchersFromNative(ctx context.Context, labelMatchers []string) (LabelMatchers, diag.Diagnostics) {
	var l LabelMatchers = make([]LabelMatcher, 0, len(labelMatchers))
	var d = diag.Diagnostics{}
	for _, v := range labelMatchers {
		elements := strings.Split(v, ",")
		var key string
		var values []string

		for _, e := range elements {
			parts := strings.Split(e, "=")
			if len(parts) != 2 {
				d.AddError("Error parsing label matcher", "invalid label matcher")
				return nil, d
			}

			if key == "" {
				key = parts[0]
			}

			values = append(values, parts[1])
		}

		val, d := types.ListValueFrom(ctx, types.StringType, values)
		if d.HasError() {
			return nil, d
		}

		l = append(l, LabelMatcher{
			Key:    types.StringValue(key),
			Values: val,
		})
	}

	return l, nil
}

func (r *FederatedGraphResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *FederatedGraphModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// We fetch the FederatedGraph list to get the requested FederatedGraph, as we don't have a direct read endpoint.
	ns, err := r.client.GetFederatedGraphs(ctx, &connect.Request[platformv1.GetFederatedGraphsRequest]{
		Msg: &platformv1.GetFederatedGraphsRequest{
			Namespace: data.Namespace.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading federated graph", err.Error())
		return
	}

	if ns.Msg.GetResponse().Code != common.EnumStatusCode_OK {
		resp.Diagnostics.AddError("Error fetching federated graph list", ns.Msg.GetResponse().GetDetails())
		return
	}

	var current *FederatedGraphModel
	for _, n := range ns.Msg.Graphs {
		if n.Id == data.Id.ValueString() {
			var admissionWebhookUrl *string
			if n.AdmissionWebhookUrl != nil && *(n.AdmissionWebhookUrl) != "" {
				admissionWebhookUrl = n.AdmissionWebhookUrl
			}

			labelMatchers, d := MapLabelMatchersFromNative(ctx, n.LabelMatchers)
			if d.HasError() {
				resp.Diagnostics.Append(d...)
				return
			}

			current = &FederatedGraphModel{
				Id:                  types.StringValue(n.Id),
				Name:                types.StringValue(n.Name),
				Namespace:           types.StringValue(n.Namespace),
				RoutingUrl:          types.StringValue(n.RoutingURL),
				AdmissionWebhookUrl: types.StringPointerValue(admissionWebhookUrl),
				LabelMatchers:       labelMatchers,
			}
			continue
		}
	}

	if current == nil {
		resp.Diagnostics.AddError("Error reading federated graph", "federated graph not found")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &current)...)
}

func (r *FederatedGraphResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FederatedGraphModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state FederatedGraphModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	labels, d := MapLabelMatchersToNative(ctx, plan.LabelMatchers)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	ru, err := r.client.UpdateFederatedGraph(ctx, &connect.Request[platformv1.UpdateFederatedGraphRequest]{
		Msg: &platformv1.UpdateFederatedGraphRequest{
			Name:                   plan.Name.ValueString(),
			Namespace:              plan.Namespace.ValueString(),
			RoutingUrl:             plan.RoutingUrl.ValueString(),
			AdmissionWebhookSecret: plan.AdmissionWebhookSecret.ValueStringPointer(),
			AdmissionWebhookURL:    plan.AdmissionWebhookUrl.ValueStringPointer(),
			LabelMatchers:          labels,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating federated graph", err.Error())
		return
	}

	if ru.Msg.GetResponse().Code != common.EnumStatusCode_OK {
		resp.Diagnostics.AddError("Error updating federated graph", ru.Msg.GetResponse().GetDetails())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FederatedGraphResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *FederatedGraphModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteFederatedGraph(ctx, &connect.Request[platformv1.DeleteFederatedGraphRequest]{
		Msg: &platformv1.DeleteFederatedGraphRequest{
			Name:      data.Name.ValueString(),
			Namespace: data.Namespace.ValueString(),
		},
	})

	if err != nil {
		resp.Diagnostics.AddError("Error deleting federated graph", err.Error())
		return
	}
}

func (r *FederatedGraphResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
