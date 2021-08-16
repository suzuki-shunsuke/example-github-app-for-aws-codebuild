# Getting Started with Terraform

We provide Terraform Configuration to setup AWS resources quickly for trial.
_Note that this Terraform Configuration is for not production use but getting started._

## Requirement

* Git
* Terraform
* AWS Access Key
* Go
* goreleaser

## Procedure

```console
$ git clone https://github.com/suzuki-shunsuke/example-github-app-for-aws-codebuild
$ cd example-github-app-for-aws-codebuild
```

Create a Zip file.

```
$ go mod download
$ goreleaser release --rm-dist --snapshot
$ mv dist/codebuilder_linux_amd64.zip terraform
```

```
$ cd terraform
```

If you use [tfenv](https://github.com/tfutils/tfenv),

```console
$ tfenv install
```

Configure Terraform [input variables](https://www.terraform.io/docs/language/values/variables.html).

```console
$ cp terraform.tfvars.template terraform.tfvars
$ vi terraform.tfvars
```

Create resources.

```console
$ terraform apply [-refresh=false]
```

`-refresh=false` is useful to make terraform commands fast.

## Clean up

```
$ terraform destroy
```
