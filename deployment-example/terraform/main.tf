# See: https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity
module "workload_identity_existing_ksa" {
  source  = "terraform-google-modules/kubernetes-engine/google//modules/workload-identity"
  version = "~> 30.0"

  project_id   = "<your-project>"
  name         = "gcp-sa-exporter"
  cluster_name = "<your-existing-cluster>"
  location     = "<your-existing-cluster-location>"

  use_existing_k8s_sa = true # it has been set via helm chart
  annotate_k8s_sa     = false
  k8s_sa_name         = "gcp-sa-exporter"
  namespace           = "gcp-sa-exporter"
  roles = [
    "roles/iam.serviceAccountViewer",
  ]
}

# If you want to use parent-id functionality
# gcp-sa-exporter additional scope for cross projects
resource "google_organization_iam_member" "gcp-sa-exporter" {
  org_id = "<org_id>"
  for_each = toset(module.workload_identity_existing_ksa.roles)
  role   = each.value
  member = "serviceAccount:${module.workload_identity_existing_ksa.gcp_service_account_email}"
}
