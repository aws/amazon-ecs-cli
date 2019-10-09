package amimetadata

import (
	"fmt"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/amimetadata/mock/sdk"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Configurer func(ssmClient *mock_ssmiface.MockSSMAPI) *mock_ssmiface.MockSSMAPI

func TestMetadataClient_GetRecommendedECSLinuxAMI(t *testing.T) {
	tests := []struct {
		instanceType  string
		configureMock Configurer
		expectedErr   error
	}{
		{
			// validate that we use the ARM64 optimized AMI for Arm instances
			"a1.medium",
			func(ssmClient *mock_ssmiface.MockSSMAPI) *mock_ssmiface.MockSSMAPI {
				ssmClient.EXPECT().GetParameter(gomock.Any()).Do(func(input *ssm.GetParameterInput) {
					assert.Equal(t, amazonLinux2ARM64RecommendedParameterName, *input.Name)
				}).Return(emptySSMParameterOutput(), nil)
				return ssmClient
			},
			nil,
		},
		{
			// validate that we use GPU optimized AMI for GPU instances
			"p2.large",
			func(ssmClient *mock_ssmiface.MockSSMAPI) *mock_ssmiface.MockSSMAPI {
				ssmClient.EXPECT().GetParameter(gomock.Any()).Do(func(input *ssm.GetParameterInput) {
					assert.Equal(t, amazonLinux2X86GPURecommendedParameterName, *input.Name)
				}).Return(emptySSMParameterOutput(), nil)
				return ssmClient
			},
			nil,
		},
		{
			// validate that we use GPU optimized AMI for GPU instances
			"g4dn.xlarge",
			func(ssmClient *mock_ssmiface.MockSSMAPI) *mock_ssmiface.MockSSMAPI {
				ssmClient.EXPECT().GetParameter(gomock.Any()).Do(func(input *ssm.GetParameterInput) {
					assert.Equal(t, amazonLinux2X86GPURecommendedParameterName, *input.Name)
				}).Return(emptySSMParameterOutput(), nil)
				return ssmClient
			},
			nil,
		},
		{
			// validate that we use the generic AMI for other instances
			"t2.micro",
			func(ssmClient *mock_ssmiface.MockSSMAPI) *mock_ssmiface.MockSSMAPI {
				ssmClient.EXPECT().GetParameter(gomock.Any()).Do(func(input *ssm.GetParameterInput) {
					assert.Equal(t, amazonLinux2X86RecommendedParameterName, *input.Name)
				}).Return(emptySSMParameterOutput(), nil)
				return ssmClient
			},
			nil,
		},
		{
			// validate that we throw an error if the AMI is not available in a region
			"t2.micro",
			func(ssmClient *mock_ssmiface.MockSSMAPI) *mock_ssmiface.MockSSMAPI {
				ssmClient.EXPECT().GetParameter(gomock.Any()).Do(func(input *ssm.GetParameterInput) {
					assert.Equal(t, amazonLinux2X86RecommendedParameterName, *input.Name)
				}).Return(nil, awserr.New(ssm.ErrCodeParameterNotFound, "some error", nil))
				return ssmClient
			},
			errors.New(fmt.Sprintf(
				"Could not find Recommended Amazon Linux 2 AMI %s in %s; the AMI may not be supported in this region: ParameterNotFound: some error",
				amazonLinux2X86RecommendedParameterName,
				"us-east-1")),
		},
		{
			// validate that we throw unexpected errors
			"t2.micro",
			func(ssmClient *mock_ssmiface.MockSSMAPI) *mock_ssmiface.MockSSMAPI {
				ssmClient.EXPECT().GetParameter(gomock.Any()).Do(func(input *ssm.GetParameterInput) {
					assert.Equal(t, amazonLinux2X86RecommendedParameterName, *input.Name)
				}).Return(nil, errors.New("unexpected error"))
				return ssmClient
			},
			errors.New("unexpected error"),
		},
	}

	for _, test := range tests {
		m := newMockSSMAPI(t)
		test.configureMock(m)

		c := metadataClient{
			m,
			"us-east-1",
		}
		_, actualErr := c.GetRecommendedECSLinuxAMI(test.instanceType)

		if test.expectedErr == nil {
			assert.NoError(t, actualErr)
		} else {
			assert.EqualError(t, actualErr, test.expectedErr.Error())
		}
	}
}

func newMockSSMAPI(t *testing.T) *mock_ssmiface.MockSSMAPI {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	return mock_ssmiface.NewMockSSMAPI(ctrl)
}

func emptySSMParameterOutput() *ssm.GetParameterOutput {
	outputJson := "{}"
	return &ssm.GetParameterOutput{
		Parameter: &ssm.Parameter{
			Value: &outputJson,
		},
	}
}
