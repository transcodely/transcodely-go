package transcodely

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// assertEnumCoverage checks that the value-constants the SDK re-exports for a
// proto enum cover every concrete value the proto declares — i.e. every value
// except the UNSPECIFIED zero sentinel. It is the drift guard for the invariant
// stated at the top of types.go: "Re-exports of every public proto message and
// enum so callers never need to import the internal gen path."
//
// A clean build only proves the SDK references no *removed/renamed* proto value
// (those wouldn't compile). It does NOT prove completeness: a value the proto
// adds but the SDK forgets to alias compiles fine. This test closes that gap —
// if the proto gains an enum value, it fails until the SDK adds the constant.
func assertEnumCoverage[E protoreflect.Enum](t *testing.T, name string, exposed []E) {
	t.Helper()
	var zero E
	values := zero.Descriptor().Values()

	valid := map[protoreflect.EnumNumber]bool{}
	concrete := map[protoreflect.EnumNumber]protoreflect.Name{}
	for i := 0; i < values.Len(); i++ {
		v := values.Get(i)
		valid[v.Number()] = true
		if v.Number() != 0 { // skip the UNSPECIFIED zero sentinel
			concrete[v.Number()] = v.Name()
		}
	}

	got := map[protoreflect.EnumNumber]bool{}
	for _, e := range exposed {
		if !valid[e.Number()] {
			t.Errorf("%s: SDK exposes a constant (=%d) that is not a proto value", name, e.Number())
		}
		got[e.Number()] = true
	}
	for num, vname := range concrete {
		if !got[num] {
			t.Errorf("%s: proto value %s (=%d) has no SDK value-constant in types.go", name, vname, num)
		}
	}
}

// TestEnumAliasCoverage asserts every typed enum the SDK re-exports covers the
// full set of concrete proto values. Each list mirrors a const block in
// types.go (or memberships.go for MembershipRole). Adding a proto enum value
// without exporting it here fails this test.
func TestEnumAliasCoverage(t *testing.T) {
	assertEnumCoverage(t, "JobStatus", []JobStatus{
		JobStatusPending, JobStatusProbing, JobStatusProcessing, JobStatusCompleted,
		JobStatusFailed, JobStatusCanceled, JobStatusPartial, JobStatusAwaitingConfirmation,
	})
	assertEnumCoverage(t, "JobPriority", []JobPriority{
		JobPriorityEconomy, JobPriorityStandard, JobPriorityPremium,
	})
	assertEnumCoverage(t, "OutputStatus", []OutputStatus{
		OutputStatusPending, OutputStatusProcessing, OutputStatusCompleted,
		OutputStatusFailed, OutputStatusCanceled,
	})
	assertEnumCoverage(t, "OutputFormat", []OutputFormat{
		OutputFormatMP4, OutputFormatWebM, OutputFormatMKV, OutputFormatMOV,
		OutputFormatHLS, OutputFormatDASH, OutputFormatAdaptive,
	})
	assertEnumCoverage(t, "VideoCodec", []VideoCodec{
		VideoCodecH264, VideoCodecH265, VideoCodecVP9, VideoCodecAV1,
	})
	assertEnumCoverage(t, "AudioCodec", []AudioCodec{
		AudioCodecAAC, AudioCodecOpus, AudioCodecMP3,
	})
	assertEnumCoverage(t, "Container", []Container{
		ContainerMP4, ContainerWebM, ContainerMKV, ContainerTS, ContainerMOV,
	})
	assertEnumCoverage(t, "Resolution", []Resolution{
		Resolution480P, Resolution720P, Resolution1080P, Resolution1440P,
		Resolution2160P, Resolution4320P,
	})
	assertEnumCoverage(t, "QualityTier", []QualityTier{
		QualityTierEconomy, QualityTierStandard, QualityTierPremium,
	})
	assertEnumCoverage(t, "ContentType", []ContentType{
		ContentTypeFilm, ContentTypeAnimation, ContentTypeGrain, ContentTypeGaming,
		ContentTypeSports, ContentTypeStillImage,
	})
	assertEnumCoverage(t, "DeliveryFormat", []DeliveryFormat{
		DeliveryFormatProgressive, DeliveryFormatHLS, DeliveryFormatDASH, DeliveryFormatCMAF,
	})
	assertEnumCoverage(t, "BitrateMode", []BitrateMode{
		BitrateModeCRF, BitrateModeCBR, BitrateModeVBR,
	})
	assertEnumCoverage(t, "WatchEventType", []WatchEventType{
		WatchEventSnapshot, WatchEventProgress, WatchEventStatusChange,
		WatchEventCompleted, WatchEventHeartbeat,
	})
	assertEnumCoverage(t, "DRMSystem", []DRMSystem{
		DRMSystemWidevine, DRMSystemFairPlay, DRMSystemPlayReady,
	})
	assertEnumCoverage(t, "EncryptionScheme", []EncryptionScheme{
		EncryptionSchemeCENC, EncryptionSchemeCBCS,
	})
	assertEnumCoverage(t, "HDRFormat", []HDRFormat{
		HDRFormatHDR10, HDRFormatHDR10Plus, HDRFormatHLG,
		HDRFormatDolbyVision5, HDRFormatDolbyVision8,
	})
	assertEnumCoverage(t, "HDRMode", []HDRMode{
		HDRModePassthrough, HDRModeTonemap, HDRModeForce,
	})
	assertEnumCoverage(t, "ToneMapping", []ToneMapping{
		ToneMappingReinhard, ToneMappingHable, ToneMappingBT2390, ToneMappingMobius,
	})
	assertEnumCoverage(t, "ContentAwareMode", []ContentAwareMode{
		ContentAwareModePerTitle, ContentAwareModeAutoABR,
	})
	assertEnumCoverage(t, "SubtitleOperation", []SubtitleOperation{
		SubtitleOpPassthrough, SubtitleOpConvert, SubtitleOpBurnIn, SubtitleOpExtract,
	})
	assertEnumCoverage(t, "SubtitleFormat", []SubtitleFormat{
		SubtitleFormatSRT, SubtitleFormatWebVTT, SubtitleFormatTTML, SubtitleFormatASS,
	})
	assertEnumCoverage(t, "ThumbnailMode", []ThumbnailMode{
		ThumbnailModeSingle, ThumbnailModeInterval, ThumbnailModeSprite, ThumbnailModeTimestamps,
	})
	assertEnumCoverage(t, "ThumbnailFormat", []ThumbnailFormat{
		ThumbnailFormatJPEG, ThumbnailFormatPNG, ThumbnailFormatWebP,
	})
	assertEnumCoverage(t, "HLSSegmentFormat", []HLSSegmentFormat{
		HLSSegmentFMP4, HLSSegmentTS,
	})
	assertEnumCoverage(t, "HLSPlaylistType", []HLSPlaylistType{
		HLSPlaylistVOD, HLSPlaylistEvent,
	})
	assertEnumCoverage(t, "GOPAlignmentMode", []GOPAlignmentMode{
		GOPAlignmentAligned, GOPAlignmentFixed,
	})
	assertEnumCoverage(t, "OriginProvider", []OriginProvider{
		OriginProviderGCS, OriginProviderS3, OriginProviderHTTP,
		OriginProviderTranscodely, OriginProviderR2,
	})
	assertEnumCoverage(t, "OriginPermission", []OriginPermission{
		OriginPermissionRead, OriginPermissionWrite,
	})
	assertEnumCoverage(t, "OriginStatus", []OriginStatus{
		OriginStatusActive, OriginStatusFailed, OriginStatusArchived,
	})
	assertEnumCoverage(t, "R2Jurisdiction", []R2Jurisdiction{
		R2JurisdictionDefault, R2JurisdictionEU, R2JurisdictionFedRAMP,
	})
	assertEnumCoverage(t, "VideoStatus", []VideoStatus{
		VideoStatusUploading, VideoStatusProcessing, VideoStatusReady,
		VideoStatusError, VideoStatusArchived, VideoStatusDeleted,
	})
	assertEnumCoverage(t, "VideoVisibility", []VideoVisibility{
		VideoVisibilityPublic, VideoVisibilityUnlisted, VideoVisibilityPrivate,
	})

	// Enums added by the drift audit (previously missing from types.go).
	assertEnumCoverage(t, "APIKeyEnvironment", []APIKeyEnvironment{
		APIKeyEnvironmentLive, APIKeyEnvironmentTest,
	})
	assertEnumCoverage(t, "AppStatus", []AppStatus{
		AppStatusActive, AppStatusArchived,
	})
	assertEnumCoverage(t, "OrganizationStatus", []OrganizationStatus{
		OrganizationStatusActive, OrganizationStatusSuspended, OrganizationStatusDeleted,
	})
	assertEnumCoverage(t, "MembershipStatus", []MembershipStatus{
		MembershipStatusActive, MembershipStatusInvited, MembershipStatusSuspended,
	})
	assertEnumCoverage(t, "UserStatus", []UserStatus{
		UserStatusActive, UserStatusSuspended, UserStatusDeleted,
	})
	assertEnumCoverage(t, "UserApprovalStatus", []UserApprovalStatus{
		UserApprovalStatusPending, UserApprovalStatusApproved, UserApprovalStatusRejected,
	})

	// MembershipRole is exported from memberships.go, not types.go.
	assertEnumCoverage(t, "MembershipRole", []MembershipRole{
		MembershipRoleOwner, MembershipRoleAdmin, MembershipRoleMember, MembershipRoleViewer,
	})
}
