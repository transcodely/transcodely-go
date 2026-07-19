package transcodely

import v1 "github.com/transcodely/transcodely-go/internal/gen/transcodely/v1"

// Re-exports of every public proto message and enum so callers never need to
// import the internal gen path. The underlying types are protobuf messages —
// every field is accessed through generated getters (e.g. job.GetId()) which
// are nil-safe.

// ---------- Core entities ----------

type (
	Job                 = v1.Job
	JobOutput           = v1.JobOutput
	OutputVariantResult = v1.OutputVariantResult
	OutputSpec          = v1.OutputSpec
	VideoVariant        = v1.VideoVariant
	AudioTrackConfig    = v1.AudioTrackConfig
	HLSConfig           = v1.HLSConfig
	DASHConfig          = v1.DASHConfig
	SegmentConfig       = v1.SegmentConfig
	ExecutionTiming     = v1.ExecutionTiming

	PricingSnapshot        = v1.PricingSnapshot
	VariantPricingSnapshot = v1.VariantPricingSnapshot

	Video          = v1.Video
	VideoRendition = v1.VideoRendition
	UploadPart     = v1.UploadPart
	CompletedPart  = v1.CompletedPart
	UsageSummary   = v1.UsageSummary
	DailyUsage     = v1.DailyUsage

	Preset        = v1.Preset
	PresetVariant = v1.PresetVariant
	VideoSettings = v1.VideoSettings
	AudioSettings = v1.AudioSettings

	Origin           = v1.Origin
	OriginRef        = v1.OriginRef
	GcsCredentials   = v1.GcsCredentials
	S3Credentials    = v1.S3Credentials
	HttpCredentials  = v1.HttpCredentials
	GcsOriginConfig  = v1.GcsOriginConfig
	S3OriginConfig   = v1.S3OriginConfig
	HttpOriginConfig = v1.HttpOriginConfig
	R2OriginConfig   = v1.R2OriginConfig
	ValidationResult = v1.ValidationResult

	App                 = v1.App
	HostingConfig       = v1.HostingConfig
	AutoProfileDefaults = v1.AutoProfileDefaults

	APIKey = v1.APIKey

	Organization          = v1.Organization
	Membership            = v1.Membership
	MembershipWithUser    = v1.MembershipWithUser
	User                  = v1.User
	UserWithOrganizations = v1.UserWithOrganizations
	UserOrganization      = v1.UserOrganization

	HealthCheckResponse = v1.HealthCheckResponse
	ComponentHealth     = v1.ComponentHealth

	WebhookEndpoint       = v1.WebhookEndpoint
	WebhookDelivery       = v1.WebhookDelivery
	EndpointHealth        = v1.GetEndpointHealthResponse
	EndpointHealthSummary = v1.EndpointHealthSummary
	HealthBucket          = v1.HealthBucket

	ErrorDetails        = v1.ErrorDetails
	ProtoFieldViolation = v1.FieldViolation
)

// ---------- Feature configs ----------
// Output-level toggles that ride inside an OutputSpec or PresetVariant.

type (
	DRMConfig       = v1.DRMConfig
	BYOKConfig      = v1.BYOKConfig
	KeyServerConfig = v1.KeyServerConfig

	HDRConfig = v1.HDRConfig

	// Content-aware encoding is currently unavailable. The API rejects any job
	// create request that sets content_aware (per-title or auto-ABR) on an
	// output with InvalidArgument — rule "parameter_unsupported" on
	// outputs[i].content_aware — until worker support ships. These types stay
	// exported for forward compatibility. See
	// https://github.com/transcodely/api/issues/167.
	ContentAwareConfig = v1.ContentAwareConfig
	AutoABRConfig      = v1.AutoABRConfig
	ContentAnalysis    = v1.ContentAnalysis

	SubtitleTrack = v1.SubtitleTrack
	BurnInStyle   = v1.BurnInStyle

	ThumbnailSpec   = v1.ThumbnailSpec
	ThumbnailResult = v1.ThumbnailResult

	StreamingConfig = v1.StreamingConfig

	H264Options = v1.H264Options
	H265Options = v1.H265Options
	VP9Options  = v1.VP9Options
	AV1Options  = v1.AV1Options

	InputMetadata      = v1.InputMetadata
	VideoStreamInfo    = v1.VideoStreamInfo
	AudioStreamInfo    = v1.AudioStreamInfo
	SubtitleStreamInfo = v1.SubtitleStreamInfo
)

// ---------- Watch events ----------

type (
	WatchJobResponse   = v1.WatchJobResponse
	WatchVideoResponse = v1.WatchVideoResponse
)

// ---------- Param types — re-exported request messages ----------

type (
	JobCreateParams              = v1.CreateJobRequest
	JobListParams                = v1.ListJobsRequest
	VideoListParams              = v1.ListVideosRequest
	VideoUpdateParams            = v1.UpdateVideoRequest
	VideoGetStatsParams          = v1.GetStatsRequest
	VideoListTopVideosParams     = v1.ListTopVideosRequest
	PresetCreateParams           = v1.CreatePresetRequest
	PresetUpdateParams           = v1.UpdatePresetRequest
	PresetListParams             = v1.ListPresetsRequest
	OriginCreateParams           = v1.CreateOriginRequest
	OriginUpdateParams           = v1.UpdateOriginRequest
	OriginListParams             = v1.ListOriginsRequest
	AppCreateParams              = v1.CreateAppRequest
	AppUpdateParams              = v1.UpdateAppRequest
	AppListParams                = v1.ListAppsRequest
	AppUpdateHostingConfigParams = v1.UpdateHostingConfigRequest
	APIKeyCreateParams           = v1.CreateAPIKeyRequest
	APIKeyListParams             = v1.ListAPIKeysRequest
	OrgCreateParams              = v1.CreateOrganizationRequest
	OrgUpdateParams              = v1.UpdateOrganizationRequest
	OrgListParams                = v1.ListOrganizationsRequest
	MembershipListParams         = v1.ListMembershipsRequest
	UserUpdateMeParams           = v1.UpdateMeRequest
	UserListParams               = v1.ListUsersRequest
	UploadCreateParams           = v1.CreateUploadRequest
	UploadCompleteParams         = v1.CompleteUploadRequest
	VideoCreateFromUrlParams     = v1.CreateFromUrlRequest
	MultipartCreateParams        = v1.CreateMultipartUploadRequest
	MultipartPartURLsParams      = v1.GetUploadPartUrlsRequest
	MultipartCompleteParams      = v1.CompleteMultipartUploadRequest
	MultipartAbortParams         = v1.AbortMultipartUploadRequest

	WebhookEndpointCreateParams = v1.CreateWebhookEndpointRequest
	WebhookEndpointUpdateParams = v1.UpdateWebhookEndpointRequest
	WebhookEndpointListParams   = v1.ListWebhookEndpointsRequest
	WebhookDeliveryListParams   = v1.ListWebhookDeliveriesRequest
	EventListParams             = v1.ListEventsRequest
)

// ---------- Pagination wire types ----------

type (
	PaginationRequest  = v1.PaginationRequest
	PaginationResponse = v1.PaginationResponse
)

// ---------- Enums ----------
// Use the typed aliases below instead of the verbose generated constants.

type (
	JobStatus      = v1.JobStatus
	JobPriority    = v1.JobPriority
	OutputStatus   = v1.OutputStatus
	OutputFormat   = v1.OutputFormat
	WatchEventType = v1.WatchEventType

	VideoCodec     = v1.VideoCodec
	AudioCodec     = v1.AudioCodec
	Container      = v1.Container
	Resolution     = v1.Resolution
	QualityTier    = v1.QualityTier
	ContentType    = v1.ContentType
	DeliveryFormat = v1.DeliveryFormat
	BitrateMode    = v1.BitrateMode

	DRMSystem        = v1.DRMSystem
	EncryptionScheme = v1.EncryptionScheme

	HDRFormat   = v1.HDRFormat
	HDRMode     = v1.HDRMode
	ToneMapping = v1.ToneMapping

	ContentAwareMode = v1.ContentAwareMode

	SubtitleOperation = v1.SubtitleOperation
	SubtitleFormat    = v1.SubtitleFormat

	ThumbnailMode   = v1.ThumbnailMode
	ThumbnailFormat = v1.ThumbnailFormat

	HLSSegmentFormat = v1.HLSSegmentFormat
	HLSPlaylistType  = v1.HLSPlaylistType
	GOPAlignmentMode = v1.GOPAlignmentMode

	OriginProvider   = v1.OriginProvider
	OriginPermission = v1.OriginPermission
	OriginStatus     = v1.OriginStatus
	R2Jurisdiction   = v1.R2Jurisdiction

	VideoStatus     = v1.VideoStatus
	VideoVisibility = v1.VideoVisibility

	MembershipStatus = v1.MembershipStatus
	UserStatus       = v1.UserStatus
)

// JobStatus values.
const (
	JobStatusUnspecified          = v1.JobStatus_JOB_STATUS_UNSPECIFIED
	JobStatusPending              = v1.JobStatus_JOB_STATUS_PENDING
	JobStatusProbing              = v1.JobStatus_JOB_STATUS_PROBING
	JobStatusProcessing           = v1.JobStatus_JOB_STATUS_PROCESSING
	JobStatusCompleted            = v1.JobStatus_JOB_STATUS_COMPLETED
	JobStatusFailed               = v1.JobStatus_JOB_STATUS_FAILED
	JobStatusCanceled             = v1.JobStatus_JOB_STATUS_CANCELED
	JobStatusPartial              = v1.JobStatus_JOB_STATUS_PARTIAL
	JobStatusAwaitingConfirmation = v1.JobStatus_JOB_STATUS_AWAITING_CONFIRMATION
)

// JobPriority values.
const (
	JobPriorityEconomy  = v1.JobPriority_JOB_PRIORITY_ECONOMY
	JobPriorityStandard = v1.JobPriority_JOB_PRIORITY_STANDARD
	JobPriorityPremium  = v1.JobPriority_JOB_PRIORITY_PREMIUM
)

// OutputStatus values.
const (
	OutputStatusPending    = v1.OutputStatus_OUTPUT_STATUS_PENDING
	OutputStatusProcessing = v1.OutputStatus_OUTPUT_STATUS_PROCESSING
	OutputStatusCompleted  = v1.OutputStatus_OUTPUT_STATUS_COMPLETED
	OutputStatusFailed     = v1.OutputStatus_OUTPUT_STATUS_FAILED
	OutputStatusCanceled   = v1.OutputStatus_OUTPUT_STATUS_CANCELED
)

// OutputFormat values.
const (
	OutputFormatMP4      = v1.OutputFormat_OUTPUT_FORMAT_MP4
	OutputFormatWebM     = v1.OutputFormat_OUTPUT_FORMAT_WEBM
	OutputFormatMKV      = v1.OutputFormat_OUTPUT_FORMAT_MKV
	OutputFormatMOV      = v1.OutputFormat_OUTPUT_FORMAT_MOV
	OutputFormatHLS      = v1.OutputFormat_OUTPUT_FORMAT_HLS
	OutputFormatDASH     = v1.OutputFormat_OUTPUT_FORMAT_DASH
	OutputFormatAdaptive = v1.OutputFormat_OUTPUT_FORMAT_ADAPTIVE
)

// VideoCodec values.
const (
	VideoCodecH264 = v1.VideoCodec_VIDEO_CODEC_H264
	VideoCodecH265 = v1.VideoCodec_VIDEO_CODEC_H265
	VideoCodecVP9  = v1.VideoCodec_VIDEO_CODEC_VP9
	VideoCodecAV1  = v1.VideoCodec_VIDEO_CODEC_AV1
)

// AudioCodec values.
const (
	AudioCodecAAC  = v1.AudioCodec_AUDIO_CODEC_AAC
	AudioCodecOpus = v1.AudioCodec_AUDIO_CODEC_OPUS
	AudioCodecMP3  = v1.AudioCodec_AUDIO_CODEC_MP3
)

// Container values.
const (
	ContainerMP4  = v1.Container_CONTAINER_MP4
	ContainerWebM = v1.Container_CONTAINER_WEBM
	ContainerMKV  = v1.Container_CONTAINER_MKV
	ContainerTS   = v1.Container_CONTAINER_TS
	ContainerMOV  = v1.Container_CONTAINER_MOV
)

// Resolution values.
const (
	Resolution480P  = v1.Resolution_RESOLUTION_480P
	Resolution720P  = v1.Resolution_RESOLUTION_720P
	Resolution1080P = v1.Resolution_RESOLUTION_1080P
	Resolution1440P = v1.Resolution_RESOLUTION_1440P
	Resolution2160P = v1.Resolution_RESOLUTION_2160P
	Resolution4320P = v1.Resolution_RESOLUTION_4320P
)

// QualityTier values.
const (
	QualityTierEconomy  = v1.QualityTier_QUALITY_TIER_ECONOMY
	QualityTierStandard = v1.QualityTier_QUALITY_TIER_STANDARD
	QualityTierPremium  = v1.QualityTier_QUALITY_TIER_PREMIUM
)

// ContentType values (encoder hint — what kind of source you're feeding it).
const (
	ContentTypeFilm       = v1.ContentType_CONTENT_TYPE_FILM
	ContentTypeAnimation  = v1.ContentType_CONTENT_TYPE_ANIMATION
	ContentTypeGrain      = v1.ContentType_CONTENT_TYPE_GRAIN
	ContentTypeGaming     = v1.ContentType_CONTENT_TYPE_GAMING
	ContentTypeSports     = v1.ContentType_CONTENT_TYPE_SPORTS
	ContentTypeStillImage = v1.ContentType_CONTENT_TYPE_STILLIMAGE
)

// DeliveryFormat values.
const (
	DeliveryFormatProgressive = v1.DeliveryFormat_DELIVERY_FORMAT_PROGRESSIVE
	DeliveryFormatHLS         = v1.DeliveryFormat_DELIVERY_FORMAT_HLS
	DeliveryFormatDASH        = v1.DeliveryFormat_DELIVERY_FORMAT_DASH
	DeliveryFormatCMAF        = v1.DeliveryFormat_DELIVERY_FORMAT_CMAF
)

// BitrateMode values.
const (
	BitrateModeCRF = v1.BitrateMode_BITRATE_MODE_CRF
	BitrateModeCBR = v1.BitrateMode_BITRATE_MODE_CBR
	BitrateModeVBR = v1.BitrateMode_BITRATE_MODE_VBR
)

// WatchEventType values.
const (
	WatchEventSnapshot     = v1.WatchEventType_WATCH_EVENT_TYPE_SNAPSHOT
	WatchEventProgress     = v1.WatchEventType_WATCH_EVENT_TYPE_PROGRESS
	WatchEventStatusChange = v1.WatchEventType_WATCH_EVENT_TYPE_STATUS_CHANGE
	WatchEventCompleted    = v1.WatchEventType_WATCH_EVENT_TYPE_COMPLETED
	WatchEventHeartbeat    = v1.WatchEventType_WATCH_EVENT_TYPE_HEARTBEAT
)

// DRMSystem values.
const (
	DRMSystemWidevine  = v1.DRMSystem_DRM_SYSTEM_WIDEVINE
	DRMSystemFairPlay  = v1.DRMSystem_DRM_SYSTEM_FAIRPLAY
	DRMSystemPlayReady = v1.DRMSystem_DRM_SYSTEM_PLAYREADY
)

// EncryptionScheme values.
const (
	EncryptionSchemeCENC = v1.EncryptionScheme_ENCRYPTION_SCHEME_CENC
	EncryptionSchemeCBCS = v1.EncryptionScheme_ENCRYPTION_SCHEME_CBCS
)

// HDRFormat values.
const (
	HDRFormatHDR10        = v1.HDRFormat_HDR_FORMAT_HDR10
	HDRFormatHDR10Plus    = v1.HDRFormat_HDR_FORMAT_HDR10_PLUS
	HDRFormatHLG          = v1.HDRFormat_HDR_FORMAT_HLG
	HDRFormatDolbyVision5 = v1.HDRFormat_HDR_FORMAT_DOLBY_VISION_5
	HDRFormatDolbyVision8 = v1.HDRFormat_HDR_FORMAT_DOLBY_VISION_8
)

// HDRMode values.
const (
	HDRModePassthrough = v1.HDRMode_HDR_MODE_PASSTHROUGH
	HDRModeTonemap     = v1.HDRMode_HDR_MODE_TONEMAP
	HDRModeForce       = v1.HDRMode_HDR_MODE_FORCE
)

// ToneMapping values.
const (
	ToneMappingReinhard = v1.ToneMapping_TONE_MAPPING_REINHARD
	ToneMappingHable    = v1.ToneMapping_TONE_MAPPING_HABLE
	ToneMappingBT2390   = v1.ToneMapping_TONE_MAPPING_BT2390
	ToneMappingMobius   = v1.ToneMapping_TONE_MAPPING_MOBIUS
)

// ContentAwareMode values.
const (
	ContentAwareModePerTitle = v1.ContentAwareMode_CONTENT_AWARE_MODE_PER_TITLE
	ContentAwareModeAutoABR  = v1.ContentAwareMode_CONTENT_AWARE_MODE_AUTO_ABR
)

// SubtitleOperation values.
const (
	SubtitleOpPassthrough = v1.SubtitleOperation_SUBTITLE_OPERATION_PASSTHROUGH
	SubtitleOpConvert     = v1.SubtitleOperation_SUBTITLE_OPERATION_CONVERT
	SubtitleOpBurnIn      = v1.SubtitleOperation_SUBTITLE_OPERATION_BURN_IN
	SubtitleOpExtract     = v1.SubtitleOperation_SUBTITLE_OPERATION_EXTRACT
)

// SubtitleFormat values.
const (
	SubtitleFormatSRT    = v1.SubtitleFormat_SUBTITLE_FORMAT_SRT
	SubtitleFormatWebVTT = v1.SubtitleFormat_SUBTITLE_FORMAT_WEBVTT
	SubtitleFormatTTML   = v1.SubtitleFormat_SUBTITLE_FORMAT_TTML
	SubtitleFormatASS    = v1.SubtitleFormat_SUBTITLE_FORMAT_ASS
)

// ThumbnailMode values.
const (
	ThumbnailModeSingle     = v1.ThumbnailMode_THUMBNAIL_MODE_SINGLE
	ThumbnailModeInterval   = v1.ThumbnailMode_THUMBNAIL_MODE_INTERVAL
	ThumbnailModeSprite     = v1.ThumbnailMode_THUMBNAIL_MODE_SPRITE
	ThumbnailModeTimestamps = v1.ThumbnailMode_THUMBNAIL_MODE_TIMESTAMPS
	ThumbnailModeAnimated   = v1.ThumbnailMode_THUMBNAIL_MODE_ANIMATED
)

// ThumbnailFormat values.
const (
	ThumbnailFormatJPEG = v1.ThumbnailFormat_THUMBNAIL_FORMAT_JPEG
	ThumbnailFormatPNG  = v1.ThumbnailFormat_THUMBNAIL_FORMAT_PNG
	ThumbnailFormatWebP = v1.ThumbnailFormat_THUMBNAIL_FORMAT_WEBP
)

// HLSSegmentFormat values.
const (
	HLSSegmentFMP4 = v1.HLSSegmentFormat_HLS_SEGMENT_FORMAT_FMP4
	HLSSegmentTS   = v1.HLSSegmentFormat_HLS_SEGMENT_FORMAT_TS
)

// HLSPlaylistType values.
const (
	HLSPlaylistVOD   = v1.HLSPlaylistType_HLS_PLAYLIST_TYPE_VOD
	HLSPlaylistEvent = v1.HLSPlaylistType_HLS_PLAYLIST_TYPE_EVENT
)

// GOPAlignmentMode values.
const (
	GOPAlignmentAligned = v1.GOPAlignmentMode_GOP_ALIGNMENT_MODE_ALIGNED
	GOPAlignmentFixed   = v1.GOPAlignmentMode_GOP_ALIGNMENT_MODE_FIXED
)

// OriginProvider values.
const (
	OriginProviderGCS         = v1.OriginProvider_ORIGIN_PROVIDER_GCS
	OriginProviderS3          = v1.OriginProvider_ORIGIN_PROVIDER_S3
	OriginProviderHTTP        = v1.OriginProvider_ORIGIN_PROVIDER_HTTP
	OriginProviderTranscodely = v1.OriginProvider_ORIGIN_PROVIDER_TRANSCODELY
	OriginProviderR2          = v1.OriginProvider_ORIGIN_PROVIDER_R2
)

// OriginPermission values.
const (
	OriginPermissionRead  = v1.OriginPermission_ORIGIN_PERMISSION_READ
	OriginPermissionWrite = v1.OriginPermission_ORIGIN_PERMISSION_WRITE
)

// OriginStatus values.
const (
	OriginStatusActive   = v1.OriginStatus_ORIGIN_STATUS_ACTIVE
	OriginStatusFailed   = v1.OriginStatus_ORIGIN_STATUS_FAILED
	OriginStatusArchived = v1.OriginStatus_ORIGIN_STATUS_ARCHIVED
)

// R2Jurisdiction values. Cloudflare R2 data-residency locations; only valid
// together with an account ID (the endpoint form selects jurisdiction via the
// URL instead).
const (
	R2JurisdictionDefault = v1.R2Jurisdiction_R2_JURISDICTION_DEFAULT
	R2JurisdictionEU      = v1.R2Jurisdiction_R2_JURISDICTION_EU
	R2JurisdictionFedRAMP = v1.R2Jurisdiction_R2_JURISDICTION_FEDRAMP
)

// VideoStatus values.
const (
	VideoStatusUploading  = v1.VideoStatus_VIDEO_STATUS_UPLOADING
	VideoStatusProcessing = v1.VideoStatus_VIDEO_STATUS_PROCESSING
	VideoStatusReady      = v1.VideoStatus_VIDEO_STATUS_READY
	VideoStatusError      = v1.VideoStatus_VIDEO_STATUS_ERROR
	VideoStatusArchived   = v1.VideoStatus_VIDEO_STATUS_ARCHIVED
	VideoStatusDeleted    = v1.VideoStatus_VIDEO_STATUS_DELETED
)

// VideoVisibility values.
const (
	VideoVisibilityPublic   = v1.VideoVisibility_VIDEO_VISIBILITY_PUBLIC
	VideoVisibilityUnlisted = v1.VideoVisibility_VIDEO_VISIBILITY_UNLISTED
	VideoVisibilityPrivate  = v1.VideoVisibility_VIDEO_VISIBILITY_PRIVATE
)

// MembershipStatus values. Used to filter [MembershipListParams].Status.
const (
	MembershipStatusUnspecified = v1.MembershipStatus_MEMBERSHIP_STATUS_UNSPECIFIED
	MembershipStatusActive      = v1.MembershipStatus_MEMBERSHIP_STATUS_ACTIVE
	MembershipStatusInvited     = v1.MembershipStatus_MEMBERSHIP_STATUS_INVITED
	MembershipStatusSuspended   = v1.MembershipStatus_MEMBERSHIP_STATUS_SUSPENDED
)

// UserStatus values. Used to filter [UserListParams].Status.
const (
	UserStatusUnspecified = v1.UserStatus_USER_STATUS_UNSPECIFIED
	UserStatusActive      = v1.UserStatus_USER_STATUS_ACTIVE
	UserStatusSuspended   = v1.UserStatus_USER_STATUS_SUSPENDED
	UserStatusDeleted     = v1.UserStatus_USER_STATUS_DELETED
)
