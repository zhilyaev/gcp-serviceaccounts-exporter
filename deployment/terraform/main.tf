module "workload_identity_existing_ksa" {
  source  = "terraform-google-modules/kubernetes-engine/google//modules/workload-identity"
  version = "~> 30.0"


  project_id   = "<your-project>"
  name         = "gcp-sa-exporter"
  cluster_name = "<your-existing-cluster>"
  location     = "<your-existing-cluster-location>"

  use_existing_k8s_sa = true
  annotate_k8s_sa     = false
  k8s_sa_name         = "gcp-sa-exporter"
  namespace           = "gcp-sa-exporter"
  roles = [
    "roles/iam.serviceAccountViewer",
  ]
}
