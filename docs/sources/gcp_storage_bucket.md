---
title: "Source: gcp_storage_bucket - Obtain logs from GCP Storage buckets"
description: "Allows users to collect logs from Google Cloud Platform (GCP) Storage buckets."
---

# Source: gcp_storage_bucket - Obtain logs from GCP Storage buckets

A Google Cloud Platform (GCP) Storage Bucket is a public cloud storage resource available in Google Cloud Platform's (GCP) Cloud Storage. It is used to store objects, which consist of data and its descriptive metadata. Cloud Storage makes it possible to store and retrieve varying amounts of data, at any time, from anywhere on the web.

## Configuration

| Property | Description | Default |
| - |----------------------------------------------------------------------------------------------|---------------------------|
| `connection` | The connection to use to connect to the GCP account. | - |
| `bucket` | The name of the GCP Storage bucket to collect logs from. | - |
| `prefix` | The prefix to filter objects in the bucket. | Defaults to bucket root. |
| `extensions` | The file extensions to collect. | Defaults to all files. |
| `file_layout` | Regex of pattern filename layout, used to extract information such as year, month, day, etc. | Default depends on Table. |
