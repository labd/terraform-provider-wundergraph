data "local_file" "schema" {
  filename = "${path.module}/my-subgraph.graphql"
}

resource "wundergraph_federated_subgraph" "my-subgraph" {
  name        = "my.subgraph"
  namespace   = "default"
  schema      = data.local_file.schema.content
  routing_url = "https://my-subgraph.com"
  labels = {
    "some" = "label"
  }
}
