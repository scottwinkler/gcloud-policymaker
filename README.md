# gcloud-policymaker
This repo is for demonstrating how to generate iam policies for terraform deployments. It will list all resources in the state file, and all resources the be modified in the plan. Then it selects the set of minimum permissions necessary to perform `terraform apply` and spits out a least priviliged policy that can be used to generate a deployment role.

## Install
There is an internal dependency on and [terraform](https://www.terraform.io/) and [terraform-plan-parser](https://github.com/lifeomic/terraform-plan-parser). Make sure both are installed before proceeding

example usage:

```
go get
go build
./gcloud-policymaker -dir="./terraform" -permissions="permissions.json"
```

example output:

```
#############################################
  Terraform Actions to Take:
#############################################

&{read google_storage_bucket resource}
&{read google_storage_bucket_object resource}
&{update google_storage_bucket_object resource}

#############################################
  Required permissions for deployment role:
#############################################

compute.projects.get
storage.buckets.get
storage.objects.create
storage.objects.delete
storage.objects.get
```