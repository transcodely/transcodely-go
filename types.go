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
	JobFee                 = v1.JobFee

	// JobCostBreakdown itemizes a job's totals — one line per output, one per
	// fee, the minimum-charge top-up when the per-job floor raised a total, and
	// a signed adjustment carrying anything the itemized lines don't account
	// for. The lines always sum to total_estimated_cost / total_actual_cost, so
	// render a cost from these rather than from the per-output cost fields,
	// which do not add up to the bill whenever the minimum charge binds.
	JobCostBreakdown = v1.JobCostBreakdown
	JobCostLine      = v1.JobCostLine

	Video          = v1.Video
	VideoRendition = v1.VideoRendition
	VideoTextTrack = v1.VideoTextTrack
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
	AppSpend            = v1.GetSpendResponse

	// PlayerConfig configures the hosted player and embeds for every video in
	// an app; CaptionStyle is its caption-rendering half (background color and
	// opacity, text color, text-size multiplier). Update via
	// client.Apps.UpdatePlayerConfig — merge semantics: unset fields keep
	// their stored value, an empty color string clears back to the player
	// default.
	PlayerConfig = v1.PlayerConfig
	CaptionStyle = v1.CaptionStyle

	APIKey = v1.APIKey

	// Invoice is one billing period's statement; InvoiceLineItem is one line of
	// its breakdown. Read them through client.Billing — see [Billing] for why
	// an API key cannot.
	Invoice         = v1.Invoice
	InvoiceLineItem = v1.InvoiceLineItem

	// BillingProfile is the organization's payment standing: whether the
	// provider holds a chargeable payment method, and the methods it will
	// describe for display. BillingPortalSession is a short-lived link into the
	// provider's hosted portal, which is the only place a card is added or
	// replaced — card details never reach this SDK.
	//
	// A BillingPaymentMethod frequently arrives with no Brand or Last4 even
	// when the card works; render it as "Card on file" rather than treating the
	// absent digits as an error.
	BillingProfile       = v1.BillingProfile
	BillingPaymentMethod = v1.BillingPaymentMethod
	BillingPortalSession = v1.BillingPortalSession

	// Budget is the organization's monthly budget together with the spend it is
	// measured against. It is telemetry, not a control: crossing 100% sends an
	// email and nothing else. The hard cap is App.MonthlySpendLimitEur, set via
	// [Apps.SetSpendLimit].
	Budget = v1.Budget

	// OutstandingBalance is usage accrued but not yet billed, plus the
	// threshold it is measured against. Reminders start at 80% and restrict
	// nothing; only at HardStopCents — twice the threshold — are new jobs
	// refused, with error code "outstanding_balance_exceeded". Work already
	// queued keeps running and playback keeps serving.
	OutstandingBalance = v1.OutstandingBalance

	// Settlement is the statement produced by paying an outstanding balance
	// mid-cycle. Its InvoiceId names an ordinary invoice, readable through
	// [Billing.GetInvoice].
	Settlement = v1.Settlement

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

	SubtitleTrack  = v1.SubtitleTrack
	SubtitleResult = v1.SubtitleResult
	BurnInStyle    = v1.BurnInStyle

	// ChapterPoint and ChapterResult carry the opt-in auto-chapters pass over
	// generated captions (SubtitleTrack.generate_chapters). Not yet populated:
	// auto-chapters is rolling out together with the `generate` subtitle
	// operation. Exported ahead of the rollout so consumers can code against
	// the shape now.
	ChapterPoint  = v1.ChapterPoint
	ChapterResult = v1.ChapterResult

	WatermarkConfig         = v1.WatermarkConfig
	WatermarkPixelPlacement = v1.WatermarkPixelPlacement

	ThumbnailSpec   = v1.ThumbnailSpec
	ThumbnailResult = v1.ThumbnailResult

	StreamingConfig = v1.StreamingConfig

	// ClipConfig trims the input to a sub-range before encoding. It rides on a
	// JobCreateParams (Clip field) and applies job-wide: every output is
	// encoded from the clipped range, and thumbnails/sprites are computed
	// within it. Cuts are frame-accurate (outputs are always re-encoded), and
	// billing keys off the produced output duration, so clipping reduces cost.
	ClipConfig = v1.ClipConfig

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
	AppUpdatePlayerConfigParams  = v1.UpdatePlayerConfigRequest
	AppUpdateSpendLimitParams    = v1.UpdateSpendLimitRequest
	AppUpdateSpendLimitResponse  = v1.UpdateSpendLimitResponse
	AppGetSpendParams            = v1.GetSpendRequest
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

	InvoiceListParams = v1.ListInvoicesRequest

	// BudgetUpdateParams sets or clears the monthly budget. Omitting AmountEur
	// clears it — see [Billing.ClearBudget].
	BudgetUpdateParams = v1.UpdateBudgetRequest

	// SettlementResult carries both halves of a mid-cycle settlement: the
	// statement that was produced, and the balance as it now stands (zero, with
	// any admission block lifted).
	SettlementResult = v1.SettleOutstandingBalanceResponse

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
	JobStatus       = v1.JobStatus
	JobPriority     = v1.JobPriority
	JobCostLineType = v1.JobCostLineType
	OutputStatus    = v1.OutputStatus
	OutputFormat    = v1.OutputFormat
	WatchEventType  = v1.WatchEventType

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

	WatermarkAnchor = v1.WatermarkAnchor

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

	AppStatus = v1.AppStatus

	InvoiceStatus      = v1.InvoiceStatus
	InvoiceLineType    = v1.InvoiceLineType
	PaymentMethodState = v1.PaymentMethodState
	BillingStanding    = v1.BillingStanding

	TrustTier               = v1.TrustTier
	ExposureThresholdSource = v1.ExposureThresholdSource
	BillingTreatment        = v1.BillingTreatment
	DunningStage            = v1.DunningStage
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

// JobCostLineType values. MinimumCharge is the top-up the per-job floor added
// on top of the renditions; Adjustment carries any difference between the
// itemized lines and the stored total, and may be negative.
const (
	JobCostLineTypeUnspecified   = v1.JobCostLineType_JOB_COST_LINE_TYPE_UNSPECIFIED
	JobCostLineTypeOutput        = v1.JobCostLineType_JOB_COST_LINE_TYPE_OUTPUT
	JobCostLineTypeMinimumCharge = v1.JobCostLineType_JOB_COST_LINE_TYPE_MINIMUM_CHARGE
	JobCostLineTypeFee           = v1.JobCostLineType_JOB_COST_LINE_TYPE_FEE
	JobCostLineTypeAdjustment    = v1.JobCostLineType_JOB_COST_LINE_TYPE_ADJUSTMENT
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
	SubtitleOpGenerate    = v1.SubtitleOperation_SUBTITLE_OPERATION_GENERATE
)

// SubtitleFormat values.
const (
	SubtitleFormatSRT    = v1.SubtitleFormat_SUBTITLE_FORMAT_SRT
	SubtitleFormatWebVTT = v1.SubtitleFormat_SUBTITLE_FORMAT_WEBVTT
	SubtitleFormatTTML   = v1.SubtitleFormat_SUBTITLE_FORMAT_TTML
	SubtitleFormatASS    = v1.SubtitleFormat_SUBTITLE_FORMAT_ASS
)

// WatermarkAnchor values. The relative-placement anchor region a watermark is
// aligned to; wire form is lowercase (e.g. "bottom_right"). Unspecified is
// treated as bottom_right server-side.
const (
	WatermarkAnchorUnspecified  = v1.WatermarkAnchor_WATERMARK_ANCHOR_UNSPECIFIED
	WatermarkAnchorTopLeft      = v1.WatermarkAnchor_WATERMARK_ANCHOR_TOP_LEFT
	WatermarkAnchorTopCenter    = v1.WatermarkAnchor_WATERMARK_ANCHOR_TOP_CENTER
	WatermarkAnchorTopRight     = v1.WatermarkAnchor_WATERMARK_ANCHOR_TOP_RIGHT
	WatermarkAnchorMiddleLeft   = v1.WatermarkAnchor_WATERMARK_ANCHOR_MIDDLE_LEFT
	WatermarkAnchorMiddleCenter = v1.WatermarkAnchor_WATERMARK_ANCHOR_MIDDLE_CENTER
	WatermarkAnchorMiddleRight  = v1.WatermarkAnchor_WATERMARK_ANCHOR_MIDDLE_RIGHT
	WatermarkAnchorBottomLeft   = v1.WatermarkAnchor_WATERMARK_ANCHOR_BOTTOM_LEFT
	WatermarkAnchorBottomCenter = v1.WatermarkAnchor_WATERMARK_ANCHOR_BOTTOM_CENTER
	WatermarkAnchorBottomRight  = v1.WatermarkAnchor_WATERMARK_ANCHOR_BOTTOM_RIGHT
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

// AppStatus values. Suspended is a reversible platform-admin cutoff: API keys
// scoped to the app are rejected with 403 until it is active again, but none of
// the app's resources are removed. Archived is the terminal state.
const (
	AppStatusUnspecified = v1.AppStatus_APP_STATUS_UNSPECIFIED
	AppStatusActive      = v1.AppStatus_APP_STATUS_ACTIVE
	AppStatusArchived    = v1.AppStatus_APP_STATUS_ARCHIVED
	AppStatusSuspended   = v1.AppStatus_APP_STATUS_SUSPENDED
)

// InvoiceStatus values. Draft appears only on the upcoming invoice; ListInvoices
// returns open and later states.
const (
	InvoiceStatusUnspecified   = v1.InvoiceStatus_INVOICE_STATUS_UNSPECIFIED
	InvoiceStatusDraft         = v1.InvoiceStatus_INVOICE_STATUS_DRAFT
	InvoiceStatusOpen          = v1.InvoiceStatus_INVOICE_STATUS_OPEN
	InvoiceStatusPaid          = v1.InvoiceStatus_INVOICE_STATUS_PAID
	InvoiceStatusVoid          = v1.InvoiceStatus_INVOICE_STATUS_VOID
	InvoiceStatusUncollectible = v1.InvoiceStatus_INVOICE_STATUS_UNCOLLECTIBLE
)

// InvoiceLineType values. Adjustment carries the residue that reconciles the
// itemized lines with the amount charged, and is routinely negative.
const (
	InvoiceLineTypeUnspecified = v1.InvoiceLineType_INVOICE_LINE_TYPE_UNSPECIFIED
	InvoiceLineTypeUsage       = v1.InvoiceLineType_INVOICE_LINE_TYPE_USAGE
	InvoiceLineTypeFee         = v1.InvoiceLineType_INVOICE_LINE_TYPE_FEE
	InvoiceLineTypeMinCharge   = v1.InvoiceLineType_INVOICE_LINE_TYPE_MIN_CHARGE
	InvoiceLineTypeAdjustment  = v1.InvoiceLineType_INVOICE_LINE_TYPE_ADJUSTMENT
)

// PaymentMethodState values. OnFile means the payment provider holds a method
// it can charge — including a method whose card metadata it declines to expose.
const (
	PaymentMethodStateUnspecified = v1.PaymentMethodState_PAYMENT_METHOD_STATE_UNSPECIFIED
	PaymentMethodStateNone        = v1.PaymentMethodState_PAYMENT_METHOD_STATE_NONE
	PaymentMethodStateOnFile      = v1.PaymentMethodState_PAYMENT_METHOD_STATE_ON_FILE
)

// BillingStanding values. Derived from the organization's billing facts, never
// assigned — there is no API to set one. Free and Delinquent both resolve usage
// limits to the free tier's, but Delinquent still bills; neither is a
// suspension. An organization exempt from the payment-method requirement always
// reports Active.
const (
	BillingStandingUnspecified = v1.BillingStanding_BILLING_STANDING_UNSPECIFIED
	BillingStandingFree        = v1.BillingStanding_BILLING_STANDING_FREE
	BillingStandingActive      = v1.BillingStanding_BILLING_STANDING_ACTIVE
	BillingStandingGrace       = v1.BillingStanding_BILLING_STANDING_GRACE
	BillingStandingDelinquent  = v1.BillingStanding_BILLING_STANDING_DELINQUENT
)

// TrustTier values. Derived from how many statements the organization has paid,
// never assigned, and it only ever moves upward. Proven imposes no threshold of
// its own — the account falls through to whatever ceiling its plan sets, which
// for most is none.
const (
	TrustTierUnspecified = v1.TrustTier_TRUST_TIER_UNSPECIFIED
	TrustTierNew         = v1.TrustTier_TRUST_TIER_NEW
	TrustTierEstablished = v1.TrustTier_TRUST_TIER_ESTABLISHED
	TrustTierProven      = v1.TrustTier_TRUST_TIER_PROVEN
)

// ExposureThresholdSource values — which rung supplied
// OutstandingBalance.ThresholdCents. OrgPlan joins by MINIMUM, so a plan can
// only tighten what the tier or an operator allowed. Unbounded means no
// threshold applies: no reminders, and admission is never refused for this
// reason.
const (
	ExposureThresholdSourceUnspecified     = v1.ExposureThresholdSource_EXPOSURE_THRESHOLD_SOURCE_UNSPECIFIED
	ExposureThresholdSourceOverride        = v1.ExposureThresholdSource_EXPOSURE_THRESHOLD_SOURCE_OVERRIDE
	ExposureThresholdSourceTrustTier       = v1.ExposureThresholdSource_EXPOSURE_THRESHOLD_SOURCE_TRUST_TIER
	ExposureThresholdSourceOrgPlan         = v1.ExposureThresholdSource_EXPOSURE_THRESHOLD_SOURCE_ORG_PLAN
	ExposureThresholdSourcePlatformDefault = v1.ExposureThresholdSource_EXPOSURE_THRESHOLD_SOURCE_PLATFORM_DEFAULT
	ExposureThresholdSourceUnbounded       = v1.ExposureThresholdSource_EXPOSURE_THRESHOLD_SOURCE_UNBOUNDED
)

// BillingTreatment values — how automation may treat an organization. Assigned
// by an operator, never derived, and orthogonal to [BillingStanding]: a trusted
// org can report Delinquent and still be untouched by every automated path.
//
// Admin surfaces only. Organization.BillingTreatment is not populated for
// customer-facing reads.
const (
	BillingTreatmentUnspecified = v1.BillingTreatment_BILLING_TREATMENT_UNSPECIFIED
	BillingTreatmentNormal      = v1.BillingTreatment_BILLING_TREATMENT_NORMAL
	BillingTreatmentTrusted     = v1.BillingTreatment_BILLING_TREATMENT_TRUSTED
	BillingTreatmentExempt      = v1.BillingTreatment_BILLING_TREATMENT_EXEMPT
)

// DunningStage values — how far the consequences of an unpaid statement have
// progressed. Not a fifth [BillingStanding]: standing is payment health, this
// is what was done about it. SoftLimited refuses NEW jobs while queued work
// finishes and playback keeps serving; Suspended stops API keys and public
// playback but leaves dashboard reads and Cancel open.
//
// Admin surfaces only in this release.
const (
	DunningStageUnspecified    = v1.DunningStage_DUNNING_STAGE_UNSPECIFIED
	DunningStageNone           = v1.DunningStage_DUNNING_STAGE_NONE
	DunningStagePastDue        = v1.DunningStage_DUNNING_STAGE_PAST_DUE
	DunningStageWarned         = v1.DunningStage_DUNNING_STAGE_WARNED
	DunningStageSoftLimited    = v1.DunningStage_DUNNING_STAGE_SOFT_LIMITED
	DunningStageSuspended      = v1.DunningStage_DUNNING_STAGE_SUSPENDED
	DunningStageDeletionWarned = v1.DunningStage_DUNNING_STAGE_DELETION_WARNED
	DunningStageWrittenOff     = v1.DunningStage_DUNNING_STAGE_WRITTEN_OFF
)
