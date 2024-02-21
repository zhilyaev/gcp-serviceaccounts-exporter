# gcp-serviceAccounts-exporter

> Monitoring expired service accounts via prometheus metrics


## GCP roles

You have to grant roles/iam.serviceAccountViewer to the exporter.

 If you want to use parent-id functionality, you should make google organization iam member for the google service account.
 
You can use the [terraform example](deployment-example/terraform/main.tf)


## Deploy to K8S

You can use the [helm example values](deployment-example/helm/README.md)
