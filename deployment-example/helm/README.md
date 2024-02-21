# Helm deployment

1. Setup values (envs, and ServiceAccount)
2. Deploy it!

```bash
helm upgrade -i gcp-sa-exporter -n gcp-sa-exporter oci://ghcr.io/zhilyaev/uni --version 1.1.3 -f values.yaml  
```
