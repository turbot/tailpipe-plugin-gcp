---
organization: Turbot
category: ["public cloud"]
icon_url: "/images/plugins/turbot/gcp.svg"
brand_color: "#ea4335"
display_name: "GCP"
name: "gcp"
description: "Tailpipe plugin for obtaining and querying logs from GCP."
og_description: Query GCP logs with SQL! Open source CLI. No DB required.
og_image: "/images/plugins/turbot/gcp-social-graphic.png"
engines: ["tailpipe"]
---

# GCP + Tailpipe

[Tailpipe](https://tailpipe.io) is an open-source CLI tool that allows you to obtain logs and query then with SQL.

[GCP](https://cloud.google.com) provides on-demand cloud computing platforms and APIs to authenticated customers on a metered pay-as-you-go basis.

<!-- TODO: Insert Example -->

## Documentation

- **[Table definitions & examples →](/plugins/turbot/gcp/tables)**

## Get started

### Install

Download and install the latest GCP plugin:

```bash
tailpipe plugin install gcp
```

### Credentials

| Item        | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| ----------- |-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Credentials | When running locally, you must configure your [Application Default Credentials](https://cloud.google.com/sdk/gcloud/reference/auth/application-default). If you are running in Cloud Shell or Cloud Code, [the tool uses the credentials you provided when you logged in, and manages any authorizations required](https://cloud.google.com/docs/authentication/provide-credentials-adc#cloud-based-dev).                                                                     |
| Permissions | Assign the `Viewer` role to your user or service account. You may also need additional permissions related to IAM policies, like `pubsub.subscriptions.getIamPolicy`, `pubsub.topics.getIamPolicy`, `storage.buckets.getIamPolicy`, since these are not included in the `Viewer` role. You can grant these by creating a custom role in your project.                                                                                                                         |
| Radius      | Each connection represents a single GCP project.                                                                                                                                                                                                                                                                                                                                                                                                                              |
| Resolution  | 1. Credentials from the JSON file specified by the `credentials` parameter in your tailpipe config.<br />2. Credentials from the JSON file specified by the `GOOGLE_APPLICATION_CREDENTIALS` environment variable.<br />3. Credentials from the default JSON file location (~/.config/gcloud/application_default_credentials.json). <br />4. Credentials from [the metadata server](https://cloud.google.com/docs/authentication/application-default-credentials#attached-sa) |

### Configuration

TODO: Elaborate on the configuration requirements around `partition`, `source` and `connection`.

## Advanced configuration options

By default, the GCP plugin uses your [Application Default Credentials](https://cloud.google.com/sdk/gcloud/reference/auth/application-default) to connect to GCP. If you have not set up ADC, simply run `gcloud auth application-default login`. This command will prompt you to log in, and then will download the application default credentials to ~/.config/gcloud/application_default_credentials.json.

For users with multiple GCP projects and more complex authentication use cases, here are some examples of advanced configuration options:

### Use a service account

Generate and download a JSON key for an existing service account using: [create service account key page](https://console.cloud.google.com/apis/credentials/serviceaccountkey).

```hcl
connection "gcp" "my_other_project" {
  project     = "my-other-project"
  credentials = "/home/me/my-service-account-creds.json"
}
```

### Use impersonation access token

Generate an impersonate access token using: [gcloud CLI command](https://cloud.google.com/iam/docs/create-short-lived-credentials-direct#gcloud_2).

```hcl
connection "gcp" "my_other_project" {
  project                  = "my-other-project"
  impersonate_access_token = "ya29.c.c0ASRK0GZ7mv8lIV0iiudmiGBs9m1gqGfBYZzV...aMYJd"
}
```

### Use impersonation service account

Generate an impersonate service account key using: [gcloud CLI command](https://cloud.google.com/iam/docs/impersonating-service-accounts#gcloud_1).

```hcl
connection "gcp" "my_other_project" {
  project                     = "my-other-project"
  impersonate_service_account = "turbie"
}
```

### Specify static credentials using environment variables

```sh
export CLOUDSDK_CORE_PROJECT=myproject
export GOOGLE_CLOUD_QUOTA_PROJECT=billingproject
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/my/creds.json
```