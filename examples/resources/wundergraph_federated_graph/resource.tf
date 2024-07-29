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
