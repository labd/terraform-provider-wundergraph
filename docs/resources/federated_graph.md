---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "wundergraph_federated_graph Resource - terraform-provider-wundergraph"
subcategory: ""
description: |-
  Federated graph.
---

# wundergraph_federated_graph (Resource)

Federated graph.

## Example Usage

```terraform
resource "wundergraph_federated_graph" "my-federated-graph" {
  name        = "my.federated.graph"
  namespace   = "default"
  routing_url = "https://my-federated-graph.com"
  label_matchers = [
    {
      key    = "some"
      values = ["label"]
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the federated graph to create. It is usually in the format of <org>.<env> and is used to uniquely identify your federated graph.
- `routing_url` (String) The routing url of your router. This is the url that the router will be accessible at.

### Optional

- `admission_webhook_secret` (String, Sensitive) The admission webhook secret is used to sign requests to the webhook url.
- `admission_webhook_url` (String) The admission webhook url. This is the url that the controlplane will use to implement admission control for the federated graph.
- `label_matchers` (Attributes List) The label matcher is used to select the subgraphs to federate. (see [below for nested schema](#nestedatt--label_matchers))
- `namespace` (String) The namespace name of the federated graph.

### Read-Only

- `id` (String) Identifier

<a id="nestedatt--label_matchers"></a>
### Nested Schema for `label_matchers`

Required:

- `key` (String) The key of the label matcher.
- `values` (List of String) The key of the label matcher.
