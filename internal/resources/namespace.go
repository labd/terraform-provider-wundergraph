package resources

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	platformv1 "github.com/labd/terraform-provider-wundergraph/sdk/wg/cosmo/platform/v1"
	"github.com/labd/terraform-provider-wundergraph/sdk/wg/cosmo/platform/v1/platformv1connect"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NamespaceResource{}
var _ resource.ResourceWithImportState = &NamespaceResource{}

func NewNamespaceResource() resource.Resource {
	return &NamespaceResource{}
}

// NamespaceResource defines the resource implementation.
type NamespaceResource struct {
	client platformv1connect.PlatformServiceClient
}

// NamespaceModel describes the resource data model.
type NamespaceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (r *NamespaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_namespace"
}

func (r *NamespaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Namespace resource example.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Example identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the namespace.",
				Required:            true,
			},
		},
	}
}

func (r *NamespaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NamespaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *NamespaceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	_, err := r.client.CreateNamespace(ctx, &connect.Request[platformv1.CreateNamespaceRequest]{
		Msg: &platformv1.CreateNamespaceRequest{
			Name: data.Name.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating namespace", err.Error())
		return
	}

	// We fetch the namespace list to get the requested namespace, as we don't have a direct read endpoint.
	ns, err := r.client.GetNamespaces(ctx, &connect.Request[platformv1.GetNamespacesRequest]{})
	if err != nil {
		resp.Diagnostics.AddError("Error reading namespaces", err.Error())
		return
	}

	var isFound bool
	for _, n := range ns.Msg.Namespaces {
		if n.Name == data.Name.ValueString() {
			data.Id = types.StringValue(n.Id)
			isFound = true
			continue
		}
	}

	if !isFound {
		resp.Diagnostics.AddError("Error reading namespaces", "Namespace not found")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NamespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *NamespaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// We fetch the namespace list to get the requested namespace, as we don't have a direct read endpoint.
	ns, err := r.client.GetNamespaces(ctx, &connect.Request[platformv1.GetNamespacesRequest]{})
	if err != nil {
		resp.Diagnostics.AddError("Error reading namespaces", err.Error())
		return
	}

	var current *NamespaceModel
	for _, n := range ns.Msg.Namespaces {
		if n.Id == data.Id.ValueString() {
			current = &NamespaceModel{
				Id:   types.StringValue(n.Id),
				Name: types.StringValue(n.Name),
			}
			continue
		}
	}

	if current == nil {
		resp.Diagnostics.AddError("Error reading namespaces", "Namespace not found")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &current)...)
}

func (r *NamespaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NamespaceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state NamespaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.RenameNamespace(ctx, &connect.Request[platformv1.RenameNamespaceRequest]{
		Msg: &platformv1.RenameNamespaceRequest{
			Name:    state.Name.ValueString(),
			NewName: plan.Name.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading namespaces", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NamespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *NamespaceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteNamespace(ctx, &connect.Request[platformv1.DeleteNamespaceRequest]{
		Msg: &platformv1.DeleteNamespaceRequest{
			Name: data.Name.ValueString(),
		},
	})

	if err != nil {
		resp.Diagnostics.AddError("Error deleting namespace", err.Error())
		return
	}
}

func (r *NamespaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
