variable "channels" {
  type = map(object({
    bucket_name : string,
    name : string,
  }))
  description = <<EOD
    A map FROM bucket names TO `hcp_packer_channel`s.
    These are the channels that will have their assignment managed by this module.
  EOD
  nullable    = false
}

variable "defaultToUnassigned" {
  type        = bool
  description = <<EOD
    If false, buckets without an assignment will not be managed by Terraform.
    If true, any Bucket without an assignment will have their assignment set to `"none"`.
  EOD
  default     = true
  nullable    = false
}

variable "ignoreIfNotSet" {
  type        = set(string)
  description = <<EOD
    A list of bucket names. 
    Buckets in the list will be evaluated as if `defaultToUnassigned == false`,
    even if it is `true`. 
    Has no effect when `defaultToUnassigned == false`.
  EOD
  default     = []
  nullable    = false
}

variable "errorIfNotSet" {
  type        = set(string)
  description = <<EOD
    A list of bucket names. 
    Buckets in the list will throw an error if an assignment isnt provided.
    Takes precedence over behavior from `ignoreIfNotSet` and `defaultToUnassigned`.
  EOD
  default     = []
  nullable    = false
}

variable "channelLinks" {
  type        = map(string)
  description = <<EOD
    A map FROM bucket name TO channel name (of the channel to fetch the version
    identifier from)
    Automatically assigns the version from another channel in the bucket to
    the bucket's channel in `channels`.
  EOD
  default     = {}
  nullable    = false
}

variable "explicitAssignments" {
  type        = map(string)
  description = <<EOD
    A map FROM bucket name TO version fingerprint
    Explicit version fingerprint assignments for buckets. If a bucket is present in
    this map and in `channelLinks`, the value provided here will be used instead.
  EOD
  default     = {}
  nullable    = false
}

variable "errorOnInvalidBucket" {
  type        = bool
  description = <<EOD
    If true, buckets present in `errorIfNotSet`, `explicitAssignments` or 
    `channelLinks` that are not present in `channels` will generate an error.
    If false, the invalid buckets will be ignored.
  EOD
  default     = true
  nullable    = false
}