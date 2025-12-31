Optional Inputs
These variables have default values and don't have to be set to use this module. You may set these variables to override their default values. This module has no required variables.

acceleration_status string
Description: (Optional) Sets the accelerate configuration of an existing bucket. Can be Enabled or Suspended.

Default: null

access_log_delivery_policy_source_accounts list(string)
Description: (Optional) List of AWS Account IDs should be allowed to deliver access logs to this bucket.

Default: []

access_log_delivery_policy_source_buckets list(string)
Description: (Optional) List of S3 bucket ARNs which should be allowed to deliver access logs to this bucket.

Default: []

acl string
Description: (Optional) The canned ACL to apply. Conflicts with `grant`

Default: null

allowed_kms_key_arn string
Description: The ARN of KMS key which should be allowed in PutObject

Default: null

analytics_configuration any
Description: Map containing bucket analytics configuration.

Default: {}

analytics_self_source_destination bool
Description: Whether or not the analytics source bucket is also the destination bucket.

Default: false

analytics_source_account_id string
Description: The analytics source account id.

Default: null

analytics_source_bucket_arn string
Description: The analytics source bucket ARN.

Default: null

attach_access_log_delivery_policy bool
Description: Controls if S3 bucket should have S3 access log delivery policy attached

Default: false

attach_analytics_destination_policy bool
Description: Controls if S3 bucket should have bucket analytics destination policy attached.

Default: false

attach_deny_incorrect_encryption_headers bool
Description: Controls if S3 bucket should deny incorrect encryption headers policy attached.

Default: false

attach_deny_incorrect_kms_key_sse bool
Description: Controls if S3 bucket policy should deny usage of incorrect KMS key SSE.

Default: false

attach_deny_insecure_transport_policy bool
Description: Controls if S3 bucket should have deny non-SSL transport policy attached

Default: false

attach_deny_unencrypted_object_uploads bool
Description: Controls if S3 bucket should deny unencrypted object uploads policy attached.

Default: false

attach_elb_log_delivery_policy bool
Description: Controls if S3 bucket should have ELB log delivery policy attached

Default: false

attach_inventory_destination_policy bool
Description: Controls if S3 bucket should have bucket inventory destination policy attached.

Default: false

attach_lb_log_delivery_policy bool
Description: Controls if S3 bucket should have ALB/NLB log delivery policy attached

Default: false

attach_policy bool
Description: Controls if S3 bucket should have bucket policy attached (set to `true` to use value of `policy` as bucket policy)

Default: false

attach_public_policy bool
Description: Controls if a user defined public bucket policy will be attached (set to `false` to allow upstream to apply defaults to the bucket)

Default: true

attach_require_latest_tls_policy bool
Description: Controls if S3 bucket should require the latest version of TLS

Default: false

block_public_acls bool
Description: Whether Amazon S3 should block public ACLs for this bucket.

Default: true

block_public_policy bool
Description: Whether Amazon S3 should block public bucket policies for this bucket.

Default: true

bucket string
Description: (Optional, Forces new resource) The name of the bucket. If omitted, Terraform will assign a random, unique name.

Default: null

bucket_prefix string
Description: (Optional, Forces new resource) Creates a unique bucket name beginning with the specified prefix. Conflicts with bucket.

Default: null

control_object_ownership bool
Description: Whether to manage S3 Bucket Ownership Controls on this bucket.

Default: false

cors_rule any
Description: List of maps containing rules for Cross-Origin Resource Sharing.

Default: []

create_bucket bool
Description: Controls if S3 bucket should be created

Default: true

expected_bucket_owner string
Description: The account ID of the expected bucket owner

Default: null

force_destroy bool
Description: (Optional, Default:false ) A boolean that indicates all objects should be deleted from the bucket so that the bucket can be destroyed without error. These objects are not recoverable.

Default: false

grant any
Description: An ACL policy grant. Conflicts with `acl`

Default: []

ignore_public_acls bool
Description: Whether Amazon S3 should ignore public ACLs for this bucket.

Default: true

intelligent_tiering any
Description: Map containing intelligent tiering configuration.

Default: {}

inventory_configuration any
Description: Map containing S3 inventory configuration.

Default: {}

inventory_self_source_destination bool
Description: Whether or not the inventory source bucket is also the destination bucket.

Default: false

inventory_source_account_id string
Description: The inventory source account id.

Default: null

inventory_source_bucket_arn string
Description: The inventory source bucket ARN.

Default: null

lifecycle_rule any
Description: List of maps containing configuration of object lifecycle management.

Default: []

logging any
Description: Map containing access bucket logging configuration.

Default: {}

metric_configuration any
Description: Map containing bucket metric configuration.

Default: []

object_lock_configuration any
Description: Map containing S3 object locking configuration.

Default: {}

object_lock_enabled bool
Description: Whether S3 bucket should have an Object Lock configuration enabled.

Default: false

object_ownership string
Description: Object ownership. Valid values: BucketOwnerEnforced, BucketOwnerPreferred or ObjectWriter. 'BucketOwnerEnforced': ACLs are disabled, and the bucket owner automatically owns and has full control over every object in the bucket. 'BucketOwnerPreferred': Objects uploaded to the bucket change ownership to the bucket owner if the objects are uploaded with the bucket-owner-full-control canned ACL. 'ObjectWriter': The uploading account will own the object if the object is uploaded with the bucket-owner-full-control canned ACL.

Default: "BucketOwnerEnforced"

owner map(string)
Description: Bucket owner's display name and ID. Conflicts with `acl`

Default: {}

policy string
Description: (Optional) A valid bucket policy JSON document. Note that if the policy document is not specific enough (but still valid), Terraform may view the policy as constantly changing in a terraform plan. In this case, please make sure you use the verbose/specific version of the policy. For more information about building AWS IAM policy documents with Terraform, see the AWS IAM Policy Document Guide.

Default: null

putin_khuylo bool
Description: Do you agree that Putin doesn't respect Ukrainian sovereignty and territorial integrity? More info: https://en.wikipedia.org/wiki/Putin_khuylo!

Default: true

replication_configuration any
Description: Map containing cross-region replication configuration.

Default: {}

request_payer string
Description: (Optional) Specifies who should bear the cost of Amazon S3 data transfer. Can be either BucketOwner or Requester. By default, the owner of the S3 bucket would incur the costs of any data transfer. See Requester Pays Buckets developer guide for more information.

Default: null

restrict_public_buckets bool
Description: Whether Amazon S3 should restrict public bucket policies for this bucket.

Default: true

server_side_encryption_configuration any
Description: Map containing server-side encryption configuration.

Default: {}

tags map(string)
Description: (Optional) A mapping of tags to assign to the bucket.

Default: {}

versioning map(string)
Description: Map containing versioning configuration.

Default: {}

website any
Description: Map containing static web-site hosting or redirect configuration.

Default: {}
