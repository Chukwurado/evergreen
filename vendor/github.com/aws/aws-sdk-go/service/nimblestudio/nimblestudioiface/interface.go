// Code generated by private/model/cli/gen-api/main.go. DO NOT EDIT.

// Package nimblestudioiface provides an interface to enable mocking the AmazonNimbleStudio service client
// for testing your code.
//
// It is important to note that this interface will have breaking changes
// when the service model is updated and adds new API operations, paginators,
// and waiters.
package nimblestudioiface

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/nimblestudio"
)

// NimbleStudioAPI provides an interface to enable mocking the
// nimblestudio.NimbleStudio service client's API operation,
// paginators, and waiters. This make unit testing your code that calls out
// to the SDK's service client's calls easier.
//
// The best way to use this interface is so the SDK's service client's calls
// can be stubbed out for unit testing your code with the SDK without needing
// to inject custom request handlers into the SDK's request pipeline.
//
//    // myFunc uses an SDK service client to make a request to
//    // AmazonNimbleStudio.
//    func myFunc(svc nimblestudioiface.NimbleStudioAPI) bool {
//        // Make svc.AcceptEulas request
//    }
//
//    func main() {
//        sess := session.New()
//        svc := nimblestudio.New(sess)
//
//        myFunc(svc)
//    }
//
// In your _test.go file:
//
//    // Define a mock struct to be used in your unit tests of myFunc.
//    type mockNimbleStudioClient struct {
//        nimblestudioiface.NimbleStudioAPI
//    }
//    func (m *mockNimbleStudioClient) AcceptEulas(input *nimblestudio.AcceptEulasInput) (*nimblestudio.AcceptEulasOutput, error) {
//        // mock response/functionality
//    }
//
//    func TestMyFunc(t *testing.T) {
//        // Setup Test
//        mockSvc := &mockNimbleStudioClient{}
//
//        myfunc(mockSvc)
//
//        // Verify myFunc's functionality
//    }
//
// It is important to note that this interface will have breaking changes
// when the service model is updated and adds new API operations, paginators,
// and waiters. Its suggested to use the pattern above for testing, or using
// tooling to generate mocks to satisfy the interfaces.
type NimbleStudioAPI interface {
	AcceptEulas(*nimblestudio.AcceptEulasInput) (*nimblestudio.AcceptEulasOutput, error)
	AcceptEulasWithContext(aws.Context, *nimblestudio.AcceptEulasInput, ...request.Option) (*nimblestudio.AcceptEulasOutput, error)
	AcceptEulasRequest(*nimblestudio.AcceptEulasInput) (*request.Request, *nimblestudio.AcceptEulasOutput)

	CreateLaunchProfile(*nimblestudio.CreateLaunchProfileInput) (*nimblestudio.CreateLaunchProfileOutput, error)
	CreateLaunchProfileWithContext(aws.Context, *nimblestudio.CreateLaunchProfileInput, ...request.Option) (*nimblestudio.CreateLaunchProfileOutput, error)
	CreateLaunchProfileRequest(*nimblestudio.CreateLaunchProfileInput) (*request.Request, *nimblestudio.CreateLaunchProfileOutput)

	CreateStreamingImage(*nimblestudio.CreateStreamingImageInput) (*nimblestudio.CreateStreamingImageOutput, error)
	CreateStreamingImageWithContext(aws.Context, *nimblestudio.CreateStreamingImageInput, ...request.Option) (*nimblestudio.CreateStreamingImageOutput, error)
	CreateStreamingImageRequest(*nimblestudio.CreateStreamingImageInput) (*request.Request, *nimblestudio.CreateStreamingImageOutput)

	CreateStreamingSession(*nimblestudio.CreateStreamingSessionInput) (*nimblestudio.CreateStreamingSessionOutput, error)
	CreateStreamingSessionWithContext(aws.Context, *nimblestudio.CreateStreamingSessionInput, ...request.Option) (*nimblestudio.CreateStreamingSessionOutput, error)
	CreateStreamingSessionRequest(*nimblestudio.CreateStreamingSessionInput) (*request.Request, *nimblestudio.CreateStreamingSessionOutput)

	CreateStreamingSessionStream(*nimblestudio.CreateStreamingSessionStreamInput) (*nimblestudio.CreateStreamingSessionStreamOutput, error)
	CreateStreamingSessionStreamWithContext(aws.Context, *nimblestudio.CreateStreamingSessionStreamInput, ...request.Option) (*nimblestudio.CreateStreamingSessionStreamOutput, error)
	CreateStreamingSessionStreamRequest(*nimblestudio.CreateStreamingSessionStreamInput) (*request.Request, *nimblestudio.CreateStreamingSessionStreamOutput)

	CreateStudio(*nimblestudio.CreateStudioInput) (*nimblestudio.CreateStudioOutput, error)
	CreateStudioWithContext(aws.Context, *nimblestudio.CreateStudioInput, ...request.Option) (*nimblestudio.CreateStudioOutput, error)
	CreateStudioRequest(*nimblestudio.CreateStudioInput) (*request.Request, *nimblestudio.CreateStudioOutput)

	CreateStudioComponent(*nimblestudio.CreateStudioComponentInput) (*nimblestudio.CreateStudioComponentOutput, error)
	CreateStudioComponentWithContext(aws.Context, *nimblestudio.CreateStudioComponentInput, ...request.Option) (*nimblestudio.CreateStudioComponentOutput, error)
	CreateStudioComponentRequest(*nimblestudio.CreateStudioComponentInput) (*request.Request, *nimblestudio.CreateStudioComponentOutput)

	DeleteLaunchProfile(*nimblestudio.DeleteLaunchProfileInput) (*nimblestudio.DeleteLaunchProfileOutput, error)
	DeleteLaunchProfileWithContext(aws.Context, *nimblestudio.DeleteLaunchProfileInput, ...request.Option) (*nimblestudio.DeleteLaunchProfileOutput, error)
	DeleteLaunchProfileRequest(*nimblestudio.DeleteLaunchProfileInput) (*request.Request, *nimblestudio.DeleteLaunchProfileOutput)

	DeleteLaunchProfileMember(*nimblestudio.DeleteLaunchProfileMemberInput) (*nimblestudio.DeleteLaunchProfileMemberOutput, error)
	DeleteLaunchProfileMemberWithContext(aws.Context, *nimblestudio.DeleteLaunchProfileMemberInput, ...request.Option) (*nimblestudio.DeleteLaunchProfileMemberOutput, error)
	DeleteLaunchProfileMemberRequest(*nimblestudio.DeleteLaunchProfileMemberInput) (*request.Request, *nimblestudio.DeleteLaunchProfileMemberOutput)

	DeleteStreamingImage(*nimblestudio.DeleteStreamingImageInput) (*nimblestudio.DeleteStreamingImageOutput, error)
	DeleteStreamingImageWithContext(aws.Context, *nimblestudio.DeleteStreamingImageInput, ...request.Option) (*nimblestudio.DeleteStreamingImageOutput, error)
	DeleteStreamingImageRequest(*nimblestudio.DeleteStreamingImageInput) (*request.Request, *nimblestudio.DeleteStreamingImageOutput)

	DeleteStreamingSession(*nimblestudio.DeleteStreamingSessionInput) (*nimblestudio.DeleteStreamingSessionOutput, error)
	DeleteStreamingSessionWithContext(aws.Context, *nimblestudio.DeleteStreamingSessionInput, ...request.Option) (*nimblestudio.DeleteStreamingSessionOutput, error)
	DeleteStreamingSessionRequest(*nimblestudio.DeleteStreamingSessionInput) (*request.Request, *nimblestudio.DeleteStreamingSessionOutput)

	DeleteStudio(*nimblestudio.DeleteStudioInput) (*nimblestudio.DeleteStudioOutput, error)
	DeleteStudioWithContext(aws.Context, *nimblestudio.DeleteStudioInput, ...request.Option) (*nimblestudio.DeleteStudioOutput, error)
	DeleteStudioRequest(*nimblestudio.DeleteStudioInput) (*request.Request, *nimblestudio.DeleteStudioOutput)

	DeleteStudioComponent(*nimblestudio.DeleteStudioComponentInput) (*nimblestudio.DeleteStudioComponentOutput, error)
	DeleteStudioComponentWithContext(aws.Context, *nimblestudio.DeleteStudioComponentInput, ...request.Option) (*nimblestudio.DeleteStudioComponentOutput, error)
	DeleteStudioComponentRequest(*nimblestudio.DeleteStudioComponentInput) (*request.Request, *nimblestudio.DeleteStudioComponentOutput)

	DeleteStudioMember(*nimblestudio.DeleteStudioMemberInput) (*nimblestudio.DeleteStudioMemberOutput, error)
	DeleteStudioMemberWithContext(aws.Context, *nimblestudio.DeleteStudioMemberInput, ...request.Option) (*nimblestudio.DeleteStudioMemberOutput, error)
	DeleteStudioMemberRequest(*nimblestudio.DeleteStudioMemberInput) (*request.Request, *nimblestudio.DeleteStudioMemberOutput)

	GetEula(*nimblestudio.GetEulaInput) (*nimblestudio.GetEulaOutput, error)
	GetEulaWithContext(aws.Context, *nimblestudio.GetEulaInput, ...request.Option) (*nimblestudio.GetEulaOutput, error)
	GetEulaRequest(*nimblestudio.GetEulaInput) (*request.Request, *nimblestudio.GetEulaOutput)

	GetLaunchProfile(*nimblestudio.GetLaunchProfileInput) (*nimblestudio.GetLaunchProfileOutput, error)
	GetLaunchProfileWithContext(aws.Context, *nimblestudio.GetLaunchProfileInput, ...request.Option) (*nimblestudio.GetLaunchProfileOutput, error)
	GetLaunchProfileRequest(*nimblestudio.GetLaunchProfileInput) (*request.Request, *nimblestudio.GetLaunchProfileOutput)

	GetLaunchProfileDetails(*nimblestudio.GetLaunchProfileDetailsInput) (*nimblestudio.GetLaunchProfileDetailsOutput, error)
	GetLaunchProfileDetailsWithContext(aws.Context, *nimblestudio.GetLaunchProfileDetailsInput, ...request.Option) (*nimblestudio.GetLaunchProfileDetailsOutput, error)
	GetLaunchProfileDetailsRequest(*nimblestudio.GetLaunchProfileDetailsInput) (*request.Request, *nimblestudio.GetLaunchProfileDetailsOutput)

	GetLaunchProfileInitialization(*nimblestudio.GetLaunchProfileInitializationInput) (*nimblestudio.GetLaunchProfileInitializationOutput, error)
	GetLaunchProfileInitializationWithContext(aws.Context, *nimblestudio.GetLaunchProfileInitializationInput, ...request.Option) (*nimblestudio.GetLaunchProfileInitializationOutput, error)
	GetLaunchProfileInitializationRequest(*nimblestudio.GetLaunchProfileInitializationInput) (*request.Request, *nimblestudio.GetLaunchProfileInitializationOutput)

	GetLaunchProfileMember(*nimblestudio.GetLaunchProfileMemberInput) (*nimblestudio.GetLaunchProfileMemberOutput, error)
	GetLaunchProfileMemberWithContext(aws.Context, *nimblestudio.GetLaunchProfileMemberInput, ...request.Option) (*nimblestudio.GetLaunchProfileMemberOutput, error)
	GetLaunchProfileMemberRequest(*nimblestudio.GetLaunchProfileMemberInput) (*request.Request, *nimblestudio.GetLaunchProfileMemberOutput)

	GetStreamingImage(*nimblestudio.GetStreamingImageInput) (*nimblestudio.GetStreamingImageOutput, error)
	GetStreamingImageWithContext(aws.Context, *nimblestudio.GetStreamingImageInput, ...request.Option) (*nimblestudio.GetStreamingImageOutput, error)
	GetStreamingImageRequest(*nimblestudio.GetStreamingImageInput) (*request.Request, *nimblestudio.GetStreamingImageOutput)

	GetStreamingSession(*nimblestudio.GetStreamingSessionInput) (*nimblestudio.GetStreamingSessionOutput, error)
	GetStreamingSessionWithContext(aws.Context, *nimblestudio.GetStreamingSessionInput, ...request.Option) (*nimblestudio.GetStreamingSessionOutput, error)
	GetStreamingSessionRequest(*nimblestudio.GetStreamingSessionInput) (*request.Request, *nimblestudio.GetStreamingSessionOutput)

	GetStreamingSessionStream(*nimblestudio.GetStreamingSessionStreamInput) (*nimblestudio.GetStreamingSessionStreamOutput, error)
	GetStreamingSessionStreamWithContext(aws.Context, *nimblestudio.GetStreamingSessionStreamInput, ...request.Option) (*nimblestudio.GetStreamingSessionStreamOutput, error)
	GetStreamingSessionStreamRequest(*nimblestudio.GetStreamingSessionStreamInput) (*request.Request, *nimblestudio.GetStreamingSessionStreamOutput)

	GetStudio(*nimblestudio.GetStudioInput) (*nimblestudio.GetStudioOutput, error)
	GetStudioWithContext(aws.Context, *nimblestudio.GetStudioInput, ...request.Option) (*nimblestudio.GetStudioOutput, error)
	GetStudioRequest(*nimblestudio.GetStudioInput) (*request.Request, *nimblestudio.GetStudioOutput)

	GetStudioComponent(*nimblestudio.GetStudioComponentInput) (*nimblestudio.GetStudioComponentOutput, error)
	GetStudioComponentWithContext(aws.Context, *nimblestudio.GetStudioComponentInput, ...request.Option) (*nimblestudio.GetStudioComponentOutput, error)
	GetStudioComponentRequest(*nimblestudio.GetStudioComponentInput) (*request.Request, *nimblestudio.GetStudioComponentOutput)

	GetStudioMember(*nimblestudio.GetStudioMemberInput) (*nimblestudio.GetStudioMemberOutput, error)
	GetStudioMemberWithContext(aws.Context, *nimblestudio.GetStudioMemberInput, ...request.Option) (*nimblestudio.GetStudioMemberOutput, error)
	GetStudioMemberRequest(*nimblestudio.GetStudioMemberInput) (*request.Request, *nimblestudio.GetStudioMemberOutput)

	ListEulaAcceptances(*nimblestudio.ListEulaAcceptancesInput) (*nimblestudio.ListEulaAcceptancesOutput, error)
	ListEulaAcceptancesWithContext(aws.Context, *nimblestudio.ListEulaAcceptancesInput, ...request.Option) (*nimblestudio.ListEulaAcceptancesOutput, error)
	ListEulaAcceptancesRequest(*nimblestudio.ListEulaAcceptancesInput) (*request.Request, *nimblestudio.ListEulaAcceptancesOutput)

	ListEulaAcceptancesPages(*nimblestudio.ListEulaAcceptancesInput, func(*nimblestudio.ListEulaAcceptancesOutput, bool) bool) error
	ListEulaAcceptancesPagesWithContext(aws.Context, *nimblestudio.ListEulaAcceptancesInput, func(*nimblestudio.ListEulaAcceptancesOutput, bool) bool, ...request.Option) error

	ListEulas(*nimblestudio.ListEulasInput) (*nimblestudio.ListEulasOutput, error)
	ListEulasWithContext(aws.Context, *nimblestudio.ListEulasInput, ...request.Option) (*nimblestudio.ListEulasOutput, error)
	ListEulasRequest(*nimblestudio.ListEulasInput) (*request.Request, *nimblestudio.ListEulasOutput)

	ListEulasPages(*nimblestudio.ListEulasInput, func(*nimblestudio.ListEulasOutput, bool) bool) error
	ListEulasPagesWithContext(aws.Context, *nimblestudio.ListEulasInput, func(*nimblestudio.ListEulasOutput, bool) bool, ...request.Option) error

	ListLaunchProfileMembers(*nimblestudio.ListLaunchProfileMembersInput) (*nimblestudio.ListLaunchProfileMembersOutput, error)
	ListLaunchProfileMembersWithContext(aws.Context, *nimblestudio.ListLaunchProfileMembersInput, ...request.Option) (*nimblestudio.ListLaunchProfileMembersOutput, error)
	ListLaunchProfileMembersRequest(*nimblestudio.ListLaunchProfileMembersInput) (*request.Request, *nimblestudio.ListLaunchProfileMembersOutput)

	ListLaunchProfileMembersPages(*nimblestudio.ListLaunchProfileMembersInput, func(*nimblestudio.ListLaunchProfileMembersOutput, bool) bool) error
	ListLaunchProfileMembersPagesWithContext(aws.Context, *nimblestudio.ListLaunchProfileMembersInput, func(*nimblestudio.ListLaunchProfileMembersOutput, bool) bool, ...request.Option) error

	ListLaunchProfiles(*nimblestudio.ListLaunchProfilesInput) (*nimblestudio.ListLaunchProfilesOutput, error)
	ListLaunchProfilesWithContext(aws.Context, *nimblestudio.ListLaunchProfilesInput, ...request.Option) (*nimblestudio.ListLaunchProfilesOutput, error)
	ListLaunchProfilesRequest(*nimblestudio.ListLaunchProfilesInput) (*request.Request, *nimblestudio.ListLaunchProfilesOutput)

	ListLaunchProfilesPages(*nimblestudio.ListLaunchProfilesInput, func(*nimblestudio.ListLaunchProfilesOutput, bool) bool) error
	ListLaunchProfilesPagesWithContext(aws.Context, *nimblestudio.ListLaunchProfilesInput, func(*nimblestudio.ListLaunchProfilesOutput, bool) bool, ...request.Option) error

	ListStreamingImages(*nimblestudio.ListStreamingImagesInput) (*nimblestudio.ListStreamingImagesOutput, error)
	ListStreamingImagesWithContext(aws.Context, *nimblestudio.ListStreamingImagesInput, ...request.Option) (*nimblestudio.ListStreamingImagesOutput, error)
	ListStreamingImagesRequest(*nimblestudio.ListStreamingImagesInput) (*request.Request, *nimblestudio.ListStreamingImagesOutput)

	ListStreamingImagesPages(*nimblestudio.ListStreamingImagesInput, func(*nimblestudio.ListStreamingImagesOutput, bool) bool) error
	ListStreamingImagesPagesWithContext(aws.Context, *nimblestudio.ListStreamingImagesInput, func(*nimblestudio.ListStreamingImagesOutput, bool) bool, ...request.Option) error

	ListStreamingSessions(*nimblestudio.ListStreamingSessionsInput) (*nimblestudio.ListStreamingSessionsOutput, error)
	ListStreamingSessionsWithContext(aws.Context, *nimblestudio.ListStreamingSessionsInput, ...request.Option) (*nimblestudio.ListStreamingSessionsOutput, error)
	ListStreamingSessionsRequest(*nimblestudio.ListStreamingSessionsInput) (*request.Request, *nimblestudio.ListStreamingSessionsOutput)

	ListStreamingSessionsPages(*nimblestudio.ListStreamingSessionsInput, func(*nimblestudio.ListStreamingSessionsOutput, bool) bool) error
	ListStreamingSessionsPagesWithContext(aws.Context, *nimblestudio.ListStreamingSessionsInput, func(*nimblestudio.ListStreamingSessionsOutput, bool) bool, ...request.Option) error

	ListStudioComponents(*nimblestudio.ListStudioComponentsInput) (*nimblestudio.ListStudioComponentsOutput, error)
	ListStudioComponentsWithContext(aws.Context, *nimblestudio.ListStudioComponentsInput, ...request.Option) (*nimblestudio.ListStudioComponentsOutput, error)
	ListStudioComponentsRequest(*nimblestudio.ListStudioComponentsInput) (*request.Request, *nimblestudio.ListStudioComponentsOutput)

	ListStudioComponentsPages(*nimblestudio.ListStudioComponentsInput, func(*nimblestudio.ListStudioComponentsOutput, bool) bool) error
	ListStudioComponentsPagesWithContext(aws.Context, *nimblestudio.ListStudioComponentsInput, func(*nimblestudio.ListStudioComponentsOutput, bool) bool, ...request.Option) error

	ListStudioMembers(*nimblestudio.ListStudioMembersInput) (*nimblestudio.ListStudioMembersOutput, error)
	ListStudioMembersWithContext(aws.Context, *nimblestudio.ListStudioMembersInput, ...request.Option) (*nimblestudio.ListStudioMembersOutput, error)
	ListStudioMembersRequest(*nimblestudio.ListStudioMembersInput) (*request.Request, *nimblestudio.ListStudioMembersOutput)

	ListStudioMembersPages(*nimblestudio.ListStudioMembersInput, func(*nimblestudio.ListStudioMembersOutput, bool) bool) error
	ListStudioMembersPagesWithContext(aws.Context, *nimblestudio.ListStudioMembersInput, func(*nimblestudio.ListStudioMembersOutput, bool) bool, ...request.Option) error

	ListStudios(*nimblestudio.ListStudiosInput) (*nimblestudio.ListStudiosOutput, error)
	ListStudiosWithContext(aws.Context, *nimblestudio.ListStudiosInput, ...request.Option) (*nimblestudio.ListStudiosOutput, error)
	ListStudiosRequest(*nimblestudio.ListStudiosInput) (*request.Request, *nimblestudio.ListStudiosOutput)

	ListStudiosPages(*nimblestudio.ListStudiosInput, func(*nimblestudio.ListStudiosOutput, bool) bool) error
	ListStudiosPagesWithContext(aws.Context, *nimblestudio.ListStudiosInput, func(*nimblestudio.ListStudiosOutput, bool) bool, ...request.Option) error

	ListTagsForResource(*nimblestudio.ListTagsForResourceInput) (*nimblestudio.ListTagsForResourceOutput, error)
	ListTagsForResourceWithContext(aws.Context, *nimblestudio.ListTagsForResourceInput, ...request.Option) (*nimblestudio.ListTagsForResourceOutput, error)
	ListTagsForResourceRequest(*nimblestudio.ListTagsForResourceInput) (*request.Request, *nimblestudio.ListTagsForResourceOutput)

	PutLaunchProfileMembers(*nimblestudio.PutLaunchProfileMembersInput) (*nimblestudio.PutLaunchProfileMembersOutput, error)
	PutLaunchProfileMembersWithContext(aws.Context, *nimblestudio.PutLaunchProfileMembersInput, ...request.Option) (*nimblestudio.PutLaunchProfileMembersOutput, error)
	PutLaunchProfileMembersRequest(*nimblestudio.PutLaunchProfileMembersInput) (*request.Request, *nimblestudio.PutLaunchProfileMembersOutput)

	PutStudioMembers(*nimblestudio.PutStudioMembersInput) (*nimblestudio.PutStudioMembersOutput, error)
	PutStudioMembersWithContext(aws.Context, *nimblestudio.PutStudioMembersInput, ...request.Option) (*nimblestudio.PutStudioMembersOutput, error)
	PutStudioMembersRequest(*nimblestudio.PutStudioMembersInput) (*request.Request, *nimblestudio.PutStudioMembersOutput)

	StartStudioSSOConfigurationRepair(*nimblestudio.StartStudioSSOConfigurationRepairInput) (*nimblestudio.StartStudioSSOConfigurationRepairOutput, error)
	StartStudioSSOConfigurationRepairWithContext(aws.Context, *nimblestudio.StartStudioSSOConfigurationRepairInput, ...request.Option) (*nimblestudio.StartStudioSSOConfigurationRepairOutput, error)
	StartStudioSSOConfigurationRepairRequest(*nimblestudio.StartStudioSSOConfigurationRepairInput) (*request.Request, *nimblestudio.StartStudioSSOConfigurationRepairOutput)

	TagResource(*nimblestudio.TagResourceInput) (*nimblestudio.TagResourceOutput, error)
	TagResourceWithContext(aws.Context, *nimblestudio.TagResourceInput, ...request.Option) (*nimblestudio.TagResourceOutput, error)
	TagResourceRequest(*nimblestudio.TagResourceInput) (*request.Request, *nimblestudio.TagResourceOutput)

	UntagResource(*nimblestudio.UntagResourceInput) (*nimblestudio.UntagResourceOutput, error)
	UntagResourceWithContext(aws.Context, *nimblestudio.UntagResourceInput, ...request.Option) (*nimblestudio.UntagResourceOutput, error)
	UntagResourceRequest(*nimblestudio.UntagResourceInput) (*request.Request, *nimblestudio.UntagResourceOutput)

	UpdateLaunchProfile(*nimblestudio.UpdateLaunchProfileInput) (*nimblestudio.UpdateLaunchProfileOutput, error)
	UpdateLaunchProfileWithContext(aws.Context, *nimblestudio.UpdateLaunchProfileInput, ...request.Option) (*nimblestudio.UpdateLaunchProfileOutput, error)
	UpdateLaunchProfileRequest(*nimblestudio.UpdateLaunchProfileInput) (*request.Request, *nimblestudio.UpdateLaunchProfileOutput)

	UpdateLaunchProfileMember(*nimblestudio.UpdateLaunchProfileMemberInput) (*nimblestudio.UpdateLaunchProfileMemberOutput, error)
	UpdateLaunchProfileMemberWithContext(aws.Context, *nimblestudio.UpdateLaunchProfileMemberInput, ...request.Option) (*nimblestudio.UpdateLaunchProfileMemberOutput, error)
	UpdateLaunchProfileMemberRequest(*nimblestudio.UpdateLaunchProfileMemberInput) (*request.Request, *nimblestudio.UpdateLaunchProfileMemberOutput)

	UpdateStreamingImage(*nimblestudio.UpdateStreamingImageInput) (*nimblestudio.UpdateStreamingImageOutput, error)
	UpdateStreamingImageWithContext(aws.Context, *nimblestudio.UpdateStreamingImageInput, ...request.Option) (*nimblestudio.UpdateStreamingImageOutput, error)
	UpdateStreamingImageRequest(*nimblestudio.UpdateStreamingImageInput) (*request.Request, *nimblestudio.UpdateStreamingImageOutput)

	UpdateStudio(*nimblestudio.UpdateStudioInput) (*nimblestudio.UpdateStudioOutput, error)
	UpdateStudioWithContext(aws.Context, *nimblestudio.UpdateStudioInput, ...request.Option) (*nimblestudio.UpdateStudioOutput, error)
	UpdateStudioRequest(*nimblestudio.UpdateStudioInput) (*request.Request, *nimblestudio.UpdateStudioOutput)

	UpdateStudioComponent(*nimblestudio.UpdateStudioComponentInput) (*nimblestudio.UpdateStudioComponentOutput, error)
	UpdateStudioComponentWithContext(aws.Context, *nimblestudio.UpdateStudioComponentInput, ...request.Option) (*nimblestudio.UpdateStudioComponentOutput, error)
	UpdateStudioComponentRequest(*nimblestudio.UpdateStudioComponentInput) (*request.Request, *nimblestudio.UpdateStudioComponentOutput)
}

var _ NimbleStudioAPI = (*nimblestudio.NimbleStudio)(nil)
