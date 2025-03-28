package work

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/caraml-dev/merlin/cluster"
	clusterMock "github.com/caraml-dev/merlin/cluster/mocks"
	"github.com/caraml-dev/merlin/mlp"
	"github.com/caraml-dev/merlin/models"
	imageBuilderMock "github.com/caraml-dev/merlin/pkg/imagebuilder/mocks"
	eventMock "github.com/caraml-dev/merlin/pkg/observability/event/mocks"
	"github.com/caraml-dev/merlin/queue"
	"github.com/caraml-dev/merlin/storage/mocks"
	"github.com/caraml-dev/merlin/webhook"
	webhookMock "github.com/caraml-dev/merlin/webhook/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestExecuteDeployment(t *testing.T) {
	isDefaultTrue := true
	loggerDestinationURL := "http://logger.default"

	env := &models.Environment{
		Name:       "env1",
		Cluster:    "cluster1",
		IsDefault:  &isDefaultTrue,
		Region:     "id",
		GcpProject: "project",
		DefaultResourceRequest: &models.ResourceRequest{
			MinReplica:    0,
			MaxReplica:    1,
			CPURequest:    resource.MustParse("1"),
			MemoryRequest: resource.MustParse("1Gi"),
			GPURequest:    resource.MustParse("0"),
		},
	}

	mlpLabels := mlp.Labels{
		{Key: "key-1", Value: "value-1"},
	}

	versionLabels := models.KV{
		"key-1": "value-11",
		"key-2": "value-2",
	}

	svcMetadata := models.Metadata{
		Labels: mlp.Labels{
			{Key: "key-1", Value: "value-11"},
			{Key: "key-2", Value: "value-2"},
		},
	}

	project := mlp.Project{Name: "project", Labels: mlpLabels}
	model := &models.Model{Name: "model", Project: project, ObservabilitySupported: false}
	version := &models.Version{ID: 1, Labels: versionLabels}
	iSvcName := fmt.Sprintf("%s-%d-1", model.Name, version.ID)
	svcName := fmt.Sprintf("%s-%d-1.project.svc.cluster.local", model.Name, version.ID)
	url := fmt.Sprintf("%s-%d-1.example.com", model.Name, version.ID)

	tests := []struct {
		name              string
		endpoint          *models.VersionEndpoint
		model             *models.Model
		version           *models.Version
		deployErr         error
		deploymentStorage func() *mocks.DeploymentStorage
		storage           func() *mocks.VersionEndpointStorage
		controller        func() *clusterMock.Controller
		imageBuilder      func() *imageBuilderMock.ImageBuilder
		webhook           func() *webhookMock.Client
		eventProducer     *eventMock.EventProducer
	}{
		{
			name:    "Success: Default",
			model:   model,
			version: version,
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				Namespace:       project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := createDefaultMockDeploymentStorage()
				mockStorage.On("OnDeploymentSuccess", mock.Anything).Return(nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:     env,
					EnvironmentName: env.Name,
					ResourceRequest: env.DefaultResourceRequest,
					VersionID:       version.ID,
					Namespace:       project.Name,
				}, nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						URL:         url,
						Metadata:    svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerWebhooks", mock.Anything, webhook.OnVersionEndpointDeployed, mock.Anything).Return(nil)
				return w
			},
			eventProducer: func() *eventMock.EventProducer {
				eProducer := &eventMock.EventProducer{}
				eProducer.On("VersionEndpointChangeEvent", mock.Anything, mock.Anything).Return(nil)
				return eProducer
			}(),
		},
		{
			name:    "Success: Default - Model Observability Supported",
			model:   &models.Model{Name: "model", Project: project, ObservabilitySupported: true},
			version: version,
			endpoint: &models.VersionEndpoint{
				EnvironmentName:          env.Name,
				ResourceRequest:          env.DefaultResourceRequest,
				VersionID:                version.ID,
				Namespace:                project.Name,
				EnableModelObservability: true,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := createDefaultMockDeploymentStorage()
				mockStorage.On("OnDeploymentSuccess", mock.Anything).Return(nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:     env,
					EnvironmentName: env.Name,
					ResourceRequest: env.DefaultResourceRequest,
					VersionID:       version.ID,
					Namespace:       project.Name,
				}, nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						URL:         url,
						Metadata:    svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerWebhooks", mock.Anything, webhook.OnVersionEndpointDeployed, mock.Anything).Return(nil)
				return w
			},

			eventProducer: func() *eventMock.EventProducer {
				producer := &eventMock.EventProducer{}
				producer.On("VersionEndpointChangeEvent", &models.VersionEndpoint{
					EnvironmentName:          env.Name,
					ResourceRequest:          env.DefaultResourceRequest,
					VersionID:                version.ID,
					Namespace:                project.Name,
					RevisionID:               models.ID(1),
					Status:                   models.EndpointRunning,
					URL:                      fmt.Sprintf("%s-%d-1.example.com", model.Name, version.ID),
					ServiceName:              fmt.Sprintf("%s-%d-1.project.svc.cluster.local", model.Name, version.ID),
					EnableModelObservability: true,
				}, &models.Model{Name: "model", Project: project, ObservabilitySupported: true}).Return(nil)
				return producer
			}(),
		},
		{
			name:    "Success: with calling webhook",
			model:   model,
			version: version,
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				Namespace:       project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := createDefaultMockDeploymentStorage()
				mockStorage.On("OnDeploymentSuccess", mock.Anything).Return(nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:     env,
					EnvironmentName: env.Name,
					ResourceRequest: env.DefaultResourceRequest,
					VersionID:       version.ID,
					Namespace:       project.Name,
				}, nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						URL:         url,
						Metadata:    svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerWebhooks", mock.Anything, webhook.OnVersionEndpointDeployed, mock.Anything).Return(nil)
				return w
			},
			eventProducer: func() *eventMock.EventProducer {
				eProducer := &eventMock.EventProducer{}
				eProducer.On("VersionEndpointChangeEvent", mock.Anything, mock.Anything).Return(nil)
				return eProducer
			}(),
		},
		{
			name:    "Success eventhough error when produce event",
			model:   &models.Model{Name: "model", Project: project, ObservabilitySupported: true},
			version: version,
			endpoint: &models.VersionEndpoint{
				EnvironmentName:          env.Name,
				ResourceRequest:          env.DefaultResourceRequest,
				VersionID:                version.ID,
				Namespace:                project.Name,
				EnableModelObservability: true,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := createDefaultMockDeploymentStorage()
				mockStorage.On("OnDeploymentSuccess", mock.Anything).Return(nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:     env,
					EnvironmentName: env.Name,
					ResourceRequest: env.DefaultResourceRequest,
					VersionID:       version.ID,
					Namespace:       project.Name,
				}, nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						URL:         url,
						Metadata:    svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerWebhooks", mock.Anything, webhook.OnVersionEndpointDeployed, mock.Anything).Return(nil)
				return w
			},

			eventProducer: func() *eventMock.EventProducer {
				producer := &eventMock.EventProducer{}
				producer.On("VersionEndpointChangeEvent", &models.VersionEndpoint{
					EnvironmentName:          env.Name,
					ResourceRequest:          env.DefaultResourceRequest,
					VersionID:                version.ID,
					Namespace:                project.Name,
					RevisionID:               models.ID(1),
					Status:                   models.EndpointRunning,
					URL:                      fmt.Sprintf("%s-%d-1.example.com", model.Name, version.ID),
					ServiceName:              fmt.Sprintf("%s-%d-1.project.svc.cluster.local", model.Name, version.ID),
					EnableModelObservability: true,
				}, &models.Model{Name: "model", Project: project, ObservabilitySupported: true}).Return(fmt.Errorf("producer error"))
				return producer
			}(),
		},
		{
			name:    "Success: Latest deployment entry in storage stuck in pending",
			model:   model,
			version: version,
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				Namespace:       project.Name,
				Status:          models.EndpointPending,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("GetLatestDeployment", mock.Anything, mock.Anything).Return(
					&models.Deployment{
						ProjectID:      model.ProjectID,
						VersionModelID: model.ID,
						VersionID:      version.ID,
						Status:         models.EndpointPending,
					}, nil)
				mockStorage.On("OnDeploymentSuccess", mock.Anything).Return(nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:     env,
					EnvironmentName: env.Name,
					ResourceRequest: env.DefaultResourceRequest,
					VersionID:       version.ID,
					Namespace:       project.Name,
				}, nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						URL:         url,
						Metadata:    svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerWebhooks", mock.Anything, webhook.OnVersionEndpointDeployed, mock.Anything).Return(nil)
				return w
			},
			eventProducer: func() *eventMock.EventProducer {
				eProducer := &eventMock.EventProducer{}
				eProducer.On("VersionEndpointChangeEvent", mock.Anything, mock.Anything).Return(nil)
				return eProducer
			}(),
		},
		{
			name:    "Success: Latest deployment entry in storage not in pending state",
			model:   model,
			version: version,
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				Namespace:       project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("GetLatestDeployment", mock.Anything, mock.Anything).Return(
					&models.Deployment{
						ProjectID:      model.ProjectID,
						VersionModelID: model.ID,
						VersionID:      version.ID,
						Status:         models.EndpointRunning,
					}, nil)
				mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
				mockStorage.On("OnDeploymentSuccess", mock.Anything).Return(nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:     env,
					EnvironmentName: env.Name,
					ResourceRequest: env.DefaultResourceRequest,
					VersionID:       version.ID,
					Namespace:       project.Name,
				}, nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						URL:         url,
						Metadata:    svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerWebhooks", mock.Anything, webhook.OnVersionEndpointDeployed, mock.Anything).Return(nil)
				return w
			},
			eventProducer: func() *eventMock.EventProducer {
				eProducer := &eventMock.EventProducer{}
				eProducer.On("VersionEndpointChangeEvent", mock.Anything, mock.Anything).Return(nil)
				return eProducer
			}(),
		},
		{
			name:    "Success: Pytorch Model",
			model:   &models.Model{Name: "model", Project: project, Type: models.ModelTypePyTorch},
			version: &models.Version{ID: 1},
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				Namespace:       project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := createDefaultMockDeploymentStorage()
				mockStorage.On("OnDeploymentSuccess", mock.Anything).Return(nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:     env,
					EnvironmentName: env.Name,
					ResourceRequest: env.DefaultResourceRequest,
					VersionID:       version.ID,
					Namespace:       project.Name,
				}, nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						URL:         url,
						Metadata:    svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerWebhooks", mock.Anything, webhook.OnVersionEndpointDeployed, mock.Anything).Return(nil)
				return w
			},
			eventProducer: func() *eventMock.EventProducer {
				eProducer := &eventMock.EventProducer{}
				eProducer.On("VersionEndpointChangeEvent", mock.Anything, mock.Anything).Return(nil)
				return eProducer
			}(),
		},
		{
			name:    "Success: empty pyfunc model",
			model:   &models.Model{Name: "model", Project: project, Type: models.ModelTypePyFunc},
			version: &models.Version{ID: 1},
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				Namespace:       project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := createDefaultMockDeploymentStorage()
				mockStorage.On("OnDeploymentSuccess", mock.Anything).Return(nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:     env,
					EnvironmentName: env.Name,
					ResourceRequest: env.DefaultResourceRequest,
					VersionID:       version.ID,
					Namespace:       project.Name,
				}, nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", context.Background(), mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						URL:         url,
						Metadata:    svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				mockImgBuilder.On("BuildImage", context.Background(), project, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return("gojek/mymodel-1:latest", nil)
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerWebhooks", mock.Anything, webhook.OnVersionEndpointDeployed, mock.Anything).Return(nil)
				return w
			},
			eventProducer: func() *eventMock.EventProducer {
				eProducer := &eventMock.EventProducer{}
				eProducer.On("VersionEndpointChangeEvent", mock.Anything, mock.Anything).Return(nil)
				return eProducer
			}(),
		},
		{
			name:    "Success: pytorch model with transformer",
			model:   &models.Model{Name: "model", Project: project, Type: models.ModelTypePyTorch},
			version: &models.Version{ID: 1},
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				Namespace:       project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := createDefaultMockDeploymentStorage()
				mockStorage.On("OnDeploymentSuccess", mock.Anything).Return(nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:     env,
					EnvironmentName: env.Name,
					ResourceRequest: env.DefaultResourceRequest,
					VersionID:       version.ID,
					Namespace:       project.Name,
				}, nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", context.Background(), mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						URL:         url,
						Metadata:    svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				mockImgBuilder.On("BuildImage", context.Background(), project, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return("gojek/mymodel-1:latest", nil)
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerWebhooks", mock.Anything, webhook.OnVersionEndpointDeployed, mock.Anything).Return(nil)
				return w
			},
			eventProducer: func() *eventMock.EventProducer {
				eProducer := &eventMock.EventProducer{}
				eProducer.On("VersionEndpointChangeEvent", mock.Anything, mock.Anything).Return(nil)
				return eProducer
			}(),
		},
		{
			name:    "Success: Default With GPU",
			model:   model,
			version: version,
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: &models.ResourceRequest{
					MinReplica:    0,
					MaxReplica:    1,
					CPURequest:    resource.MustParse("1"),
					MemoryRequest: resource.MustParse("1Gi"),
					GPUName:       "NVIDIA P4",
					GPURequest:    resource.MustParse("1"),
				},
				VersionID: version.ID,
				Namespace: project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := createDefaultMockDeploymentStorage()
				mockStorage.On("OnDeploymentSuccess", mock.Anything).Return(nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:     env,
					EnvironmentName: env.Name,
					ResourceRequest: &models.ResourceRequest{
						MinReplica:    0,
						MaxReplica:    1,
						CPURequest:    resource.MustParse("1"),
						MemoryRequest: resource.MustParse("1Gi"),
						GPUName:       "NVIDIA P4",
						GPURequest:    resource.MustParse("1"),
					},
					VersionID: version.ID,
					Namespace: project.Name,
				}, nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:        iSvcName,
						Namespace:   project.Name,
						ServiceName: svcName,
						URL:         url,
						Metadata:    svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerWebhooks", mock.Anything, webhook.OnVersionEndpointDeployed, mock.Anything).Return(nil)
				return w
			},
			eventProducer: func() *eventMock.EventProducer {
				eProducer := &eventMock.EventProducer{}
				eProducer.On("VersionEndpointChangeEvent", mock.Anything, mock.Anything).Return(nil)
				return eProducer
			}(),
		},
		{
			name:      "Failed: deployment failed",
			model:     model,
			version:   version,
			deployErr: errors.New("Failed to deploy"),
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				Namespace:       project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := createDefaultMockDeploymentStorage()
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:     env,
					EnvironmentName: env.Name,
					ResourceRequest: env.DefaultResourceRequest,
					VersionID:       version.ID,
					Namespace:       project.Name,
				}, nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("Failed to deploy"))
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			eventProducer: func() *eventMock.EventProducer {
				eProducer := &eventMock.EventProducer{}
				eProducer.On("VersionEndpointChangeEvent", mock.Anything, mock.Anything).Return(nil)
				return eProducer
			}(),
		},
		{
			name:      "Failed: image builder failed",
			model:     &models.Model{Name: "model", Project: project, Type: models.ModelTypePyFunc},
			version:   version,
			deployErr: errors.New("Failed to build image"),
			endpoint: &models.VersionEndpoint{
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				Namespace:       project.Name,
			},
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := createDefaultMockDeploymentStorage()
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Save", mock.Anything).Return(nil)
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:     env,
					EnvironmentName: env.Name,
					ResourceRequest: env.DefaultResourceRequest,
					VersionID:       version.ID,
					Namespace:       project.Name,
				}, nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				mockImgBuilder.On("BuildImage", context.Background(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("Failed to build image"))
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			eventProducer: func() *eventMock.EventProducer {
				eProducer := &eventMock.EventProducer{}
				eProducer.On("VersionEndpointChangeEvent", mock.Anything, mock.Anything).Return(nil)
				return eProducer
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := tt.controller()
			controllers := map[string]cluster.Controller{env.Name: ctrl}
			imgBuilder := tt.imageBuilder()
			mockStorage := tt.storage()
			mockDeploymentStorage := tt.deploymentStorage()
			mockWebhook := tt.webhook()
			job := &queue.Job{
				Name: "job",
				Arguments: queue.Arguments{
					dataArgKey: EndpointJob{
						Endpoint: tt.endpoint,
						Version:  tt.version,
						Model:    tt.model,
						Project:  tt.model.Project,
					},
				},
			}
			svc := &ModelServiceDeployment{
				ClusterControllers:         controllers,
				ImageBuilder:               imgBuilder,
				Storage:                    mockStorage,
				DeploymentStorage:          mockDeploymentStorage,
				LoggerDestinationURL:       loggerDestinationURL,
				ObservabilityEventProducer: tt.eventProducer,
				Webhook:                    mockWebhook,
			}

			err := svc.Deploy(job)
			assert.Equal(t, tt.deployErr, err)

			if len(ctrl.ExpectedCalls) > 0 && ctrl.ExpectedCalls[0].ReturnArguments[0] != nil {
				deployedSvc := ctrl.ExpectedCalls[0].ReturnArguments[0].(*models.Service)
				assert.Equal(t, svcMetadata, deployedSvc.Metadata)
				assert.Equal(t, iSvcName, deployedSvc.Name)
			}

			mockStorage.AssertNumberOfCalls(t, "Save", 1)

			savedEndpoint := mockStorage.Calls[1].Arguments[0].(*models.VersionEndpoint)
			assert.Equal(t, tt.model.ID, savedEndpoint.VersionModelID)
			assert.Equal(t, tt.version.ID, savedEndpoint.VersionID)
			assert.Equal(t, tt.model.Project.Name, savedEndpoint.Namespace)
			assert.Equal(t, env.Name, savedEndpoint.EnvironmentName)

			if tt.endpoint.ResourceRequest != nil {
				assert.Equal(t, tt.endpoint.ResourceRequest, savedEndpoint.ResourceRequest)
			} else {
				assert.Equal(t, env.DefaultResourceRequest, savedEndpoint.ResourceRequest)
			}

			mockDeploymentStorage.AssertNumberOfCalls(t, "GetLatestDeployment", 1)
			if tt.deployErr != nil {
				mockDeploymentStorage.AssertNumberOfCalls(t, "Save", 2)
				assert.Equal(t, models.EndpointFailed, savedEndpoint.Status)
			} else {
				if tt.endpoint.Status == models.EndpointPending {
					mockDeploymentStorage.AssertNumberOfCalls(t, "Save", 0)
				} else {
					mockDeploymentStorage.AssertNumberOfCalls(t, "Save", 1)
				}
				mockDeploymentStorage.AssertNumberOfCalls(t, "OnDeploymentSuccess", 1)
				assert.Equal(t, models.EndpointRunning, savedEndpoint.Status)
				assert.Equal(t, url, savedEndpoint.URL)
				assert.Equal(t, "", savedEndpoint.InferenceServiceName)
			}
		})
	}
}

func TestExecuteRedeployment(t *testing.T) {
	isDefaultTrue := true
	loggerDestinationURL := "http://logger.default"

	env := &models.Environment{
		Name:       "env1",
		Cluster:    "cluster1",
		IsDefault:  &isDefaultTrue,
		Region:     "id",
		GcpProject: "project",
		DefaultResourceRequest: &models.ResourceRequest{
			MinReplica:    0,
			MaxReplica:    1,
			CPURequest:    resource.MustParse("1"),
			MemoryRequest: resource.MustParse("1Gi"),
			GPURequest:    resource.MustParse("0"),
		},
	}

	mlpLabels := mlp.Labels{
		{Key: "key-1", Value: "value-1"},
	}

	versionLabels := models.KV{
		"key-1": "value-11",
		"key-2": "value-2",
	}

	svcMetadata := models.Metadata{
		Labels: mlp.Labels{
			{Key: "key-1", Value: "value-11"},
			{Key: "key-2", Value: "value-2"},
		},
	}

	project := mlp.Project{Name: "project", Labels: mlpLabels}
	model := &models.Model{Name: "model", Project: project}
	version := &models.Version{ID: 1, Labels: versionLabels}

	modelSvcName := fmt.Sprintf("%s-%d-2", model.Name, version.ID)
	svcName := fmt.Sprintf("%s-%d-r2.project.svc.cluster.local", model.Name, version.ID)
	url := fmt.Sprintf("%s-%d-r2.example.com", model.Name, version.ID)

	tests := []struct {
		name                   string
		endpoint               *models.VersionEndpoint
		model                  *models.Model
		version                *models.Version
		expectedEndpointStatus models.EndpointStatus
		deployErr              error
		deploymentStorage      func() *mocks.DeploymentStorage
		storage                func() *mocks.VersionEndpointStorage
		controller             func() *clusterMock.Controller
		imageBuilder           func() *imageBuilderMock.ImageBuilder
		webhook                func() *webhookMock.Client
		eventProducer          *eventMock.EventProducer
	}{
		{
			name:    "Success: Redeploy running endpoint",
			model:   model,
			version: version,
			endpoint: &models.VersionEndpoint{
				Environment:     env,
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				RevisionID:      models.ID(1),
				Status:          models.EndpointRunning,
				Namespace:       project.Name,
			},
			expectedEndpointStatus: models.EndpointRunning,
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("GetLatestDeployment", mock.Anything, mock.Anything).Return(
					&models.Deployment{
						ProjectID:      model.ProjectID,
						VersionModelID: model.ID,
						VersionID:      version.ID,
						Status:         models.EndpointRunning,
					}, nil)
				mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
				mockStorage.On("OnDeploymentSuccess", mock.Anything).Return(nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					RevisionID:           models.ID(1),
					InferenceServiceName: fmt.Sprintf("%s-%d-1", model.Name, version.ID),
					Status:               models.EndpointRunning,
				}, nil)
				mockStorage.On("Save", &models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					RevisionID:           models.ID(2),
					InferenceServiceName: modelSvcName,
					Status:               models.EndpointRunning,
					URL:                  url,
					ServiceName:          svcName,
				}).Return(nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:            fmt.Sprintf("%s-%d-2", model.Name, version.ID),
						CurrentIsvcName: fmt.Sprintf("%s-%d-2", model.Name, version.ID),
						RevisionID:      models.ID(2),
						Namespace:       project.Name,
						ServiceName:     fmt.Sprintf("%s-%d-r2.project.svc.cluster.local", model.Name, version.ID),
						URL:             fmt.Sprintf("%s-%d-r2.example.com", model.Name, version.ID),
						Metadata:        svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerWebhooks", mock.Anything, webhook.OnVersionEndpointDeployed, mock.Anything).Return(nil)
				return w
			},
			eventProducer: func() *eventMock.EventProducer {
				eProducer := &eventMock.EventProducer{}
				eProducer.On("VersionEndpointChangeEvent", mock.Anything, mock.Anything).Return(nil)
				return eProducer
			}(),
		},
		{
			name:    "Success: Redeploy serving endpoint",
			model:   model,
			version: version,
			endpoint: &models.VersionEndpoint{
				Environment:     env,
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				RevisionID:      models.ID(1),
				Status:          models.EndpointServing,
				Namespace:       project.Name,
			},
			expectedEndpointStatus: models.EndpointServing,
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("GetLatestDeployment", mock.Anything, mock.Anything).Return(
					&models.Deployment{
						ProjectID:      model.ProjectID,
						VersionModelID: model.ID,
						VersionID:      version.ID,
						Status:         models.EndpointServing,
					}, nil)
				mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
				mockStorage.On("OnDeploymentSuccess", mock.Anything).Return(nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					RevisionID:           models.ID(1),
					InferenceServiceName: fmt.Sprintf("%s-%d-1", model.Name, version.ID),
					Status:               models.EndpointServing,
				}, nil)
				mockStorage.On("Save", &models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					RevisionID:           models.ID(2),
					InferenceServiceName: modelSvcName,
					Status:               models.EndpointServing,
					URL:                  url,
					ServiceName:          svcName,
				}).Return(nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:            fmt.Sprintf("%s-%d-2", model.Name, version.ID),
						CurrentIsvcName: fmt.Sprintf("%s-%d-2", model.Name, version.ID),
						RevisionID:      models.ID(2),
						Namespace:       project.Name,
						ServiceName:     fmt.Sprintf("%s-%d-r2.project.svc.cluster.local", model.Name, version.ID),
						URL:             fmt.Sprintf("%s-%d-r2.example.com", model.Name, version.ID),
						Metadata:        svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerWebhooks", mock.Anything, webhook.OnVersionEndpointDeployed, mock.Anything).Return(nil)
				return w
			},
			eventProducer: func() *eventMock.EventProducer {
				eProducer := &eventMock.EventProducer{}
				eProducer.On("VersionEndpointChangeEvent", mock.Anything, mock.Anything).Return(nil)
				return eProducer
			}(),
		},
		{
			name:    "Success: Redeploy failed endpoint",
			model:   model,
			version: version,
			endpoint: &models.VersionEndpoint{
				Environment:     env,
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				RevisionID:      models.ID(1),
				Status:          models.EndpointFailed,
				Namespace:       project.Name,
			},
			expectedEndpointStatus: models.EndpointRunning,
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("GetLatestDeployment", mock.Anything, mock.Anything).Return(
					&models.Deployment{
						ProjectID:      model.ProjectID,
						VersionModelID: model.ID,
						VersionID:      version.ID,
						Status:         models.EndpointFailed,
					}, nil)
				mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
				mockStorage.On("OnDeploymentSuccess", mock.Anything).Return(nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					RevisionID:           models.ID(1),
					InferenceServiceName: fmt.Sprintf("%s-%d-1", model.Name, version.ID),
					Status:               models.EndpointFailed,
				}, nil)
				mockStorage.On("Save", &models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					RevisionID:           models.ID(2),
					InferenceServiceName: modelSvcName,
					Status:               models.EndpointRunning,
					URL:                  url,
					ServiceName:          svcName,
				}).Return(nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything, mock.Anything).
					Return(&models.Service{
						Name:            fmt.Sprintf("%s-%d-2", model.Name, version.ID),
						CurrentIsvcName: fmt.Sprintf("%s-%d-2", model.Name, version.ID),
						RevisionID:      models.ID(2),
						Namespace:       project.Name,
						ServiceName:     fmt.Sprintf("%s-%d-r2.project.svc.cluster.local", model.Name, version.ID),
						URL:             fmt.Sprintf("%s-%d-r2.example.com", model.Name, version.ID),
						Metadata:        svcMetadata,
					}, nil)
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				w := webhookMock.NewClient(t)
				w.On("TriggerWebhooks", mock.Anything, webhook.OnVersionEndpointDeployed, mock.Anything).Return(nil)
				return w
			},
			eventProducer: func() *eventMock.EventProducer {
				eProducer := &eventMock.EventProducer{}
				eProducer.On("VersionEndpointChangeEvent", mock.Anything, mock.Anything).Return(nil)
				return eProducer
			}(),
		},
		{
			name:      "Failed to redeploy running endpoint",
			model:     model,
			version:   version,
			deployErr: errors.New("Failed to deploy"),
			endpoint: &models.VersionEndpoint{
				Environment:     env,
				EnvironmentName: env.Name,
				ResourceRequest: env.DefaultResourceRequest,
				VersionID:       version.ID,
				RevisionID:      models.ID(1),
				Status:          models.EndpointRunning,
				Namespace:       project.Name,
			},
			expectedEndpointStatus: models.EndpointRunning,
			deploymentStorage: func() *mocks.DeploymentStorage {
				mockStorage := &mocks.DeploymentStorage{}
				mockStorage.On("GetLatestDeployment", mock.Anything, mock.Anything).Return(
					&models.Deployment{
						ProjectID:      model.ProjectID,
						VersionModelID: model.ID,
						VersionID:      version.ID,
						Status:         models.EndpointRunning,
					}, nil)
				mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
				return mockStorage
			},
			storage: func() *mocks.VersionEndpointStorage {
				mockStorage := &mocks.VersionEndpointStorage{}
				mockStorage.On("Get", mock.Anything).Return(&models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					RevisionID:           models.ID(1),
					InferenceServiceName: fmt.Sprintf("%s-%d-1", model.Name, version.ID),
					Status:               models.EndpointRunning,
				}, nil)
				mockStorage.On("Save", &models.VersionEndpoint{
					Environment:          env,
					EnvironmentName:      env.Name,
					ResourceRequest:      env.DefaultResourceRequest,
					VersionID:            version.ID,
					Namespace:            project.Name,
					RevisionID:           models.ID(1),
					InferenceServiceName: fmt.Sprintf("%s-%d-1", model.Name, version.ID),
					Status:               models.EndpointRunning,
				}).Return(nil)
				return mockStorage
			},
			controller: func() *clusterMock.Controller {
				ctrl := &clusterMock.Controller{}
				ctrl.On("Deploy", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("Failed to deploy"))
				return ctrl
			},
			imageBuilder: func() *imageBuilderMock.ImageBuilder {
				mockImgBuilder := &imageBuilderMock.ImageBuilder{}
				return mockImgBuilder
			},
			webhook: func() *webhookMock.Client {
				return webhookMock.NewClient(t)
			},
			eventProducer: func() *eventMock.EventProducer {
				eProducer := &eventMock.EventProducer{}
				eProducer.On("VersionEndpointChangeEvent", mock.Anything, mock.Anything).Return(nil)
				return eProducer
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := tt.controller()
			controllers := map[string]cluster.Controller{env.Name: ctrl}
			imgBuilder := tt.imageBuilder()
			mockStorage := tt.storage()
			mockDeploymentStorage := tt.deploymentStorage()
			mockWebhook := tt.webhook()
			job := &queue.Job{
				Name: "job",
				Arguments: queue.Arguments{
					dataArgKey: EndpointJob{
						Endpoint: tt.endpoint,
						Version:  tt.version,
						Model:    tt.model,
						Project:  tt.model.Project,
					},
				},
			}
			svc := &ModelServiceDeployment{
				ClusterControllers:         controllers,
				ImageBuilder:               imgBuilder,
				Storage:                    mockStorage,
				DeploymentStorage:          mockDeploymentStorage,
				LoggerDestinationURL:       loggerDestinationURL,
				Webhook:                    mockWebhook,
				ObservabilityEventProducer: tt.eventProducer,
			}

			err := svc.Deploy(job)
			assert.Equal(t, tt.deployErr, err)

			if len(ctrl.ExpectedCalls) > 0 && ctrl.ExpectedCalls[0].ReturnArguments[0] != nil {
				deployedSvc := ctrl.ExpectedCalls[0].ReturnArguments[0].(*models.Service)
				assert.Equal(t, svcMetadata, deployedSvc.Metadata)
				assert.Equal(t, modelSvcName, deployedSvc.Name)
			}

			mockStorage.AssertNumberOfCalls(t, "Save", 1)

			savedEndpoint := mockStorage.Calls[1].Arguments[0].(*models.VersionEndpoint)
			assert.Equal(t, tt.model.ID, savedEndpoint.VersionModelID)
			assert.Equal(t, tt.version.ID, savedEndpoint.VersionID)
			assert.Equal(t, tt.model.Project.Name, savedEndpoint.Namespace)
			assert.Equal(t, env.Name, savedEndpoint.EnvironmentName)

			if tt.endpoint.ResourceRequest != nil {
				assert.Equal(t, tt.endpoint.ResourceRequest, savedEndpoint.ResourceRequest)
			} else {
				assert.Equal(t, env.DefaultResourceRequest, savedEndpoint.ResourceRequest)
			}

			assert.Equal(t, tt.expectedEndpointStatus, savedEndpoint.Status)

			mockDeploymentStorage.AssertNumberOfCalls(t, "GetLatestDeployment", 1)
			if tt.deployErr == nil {
				if tt.endpoint.Status == models.EndpointPending {
					mockDeploymentStorage.AssertNumberOfCalls(t, "Save", 0)
				} else {
					mockDeploymentStorage.AssertNumberOfCalls(t, "Save", 1)
				}
				mockDeploymentStorage.AssertNumberOfCalls(t, "OnDeploymentSuccess", 1)
				assert.Equal(t, url, savedEndpoint.URL)
				assert.Equal(t, modelSvcName, savedEndpoint.InferenceServiceName)
			} else {
				mockDeploymentStorage.AssertNumberOfCalls(t, "Save", 2)
			}
		})
	}
}

func createDefaultMockDeploymentStorage() *mocks.DeploymentStorage {
	mockStorage := &mocks.DeploymentStorage{}
	mockStorage.On("GetLatestDeployment", mock.Anything, mock.Anything).Return(nil, gorm.ErrRecordNotFound)
	mockStorage.On("Save", mock.Anything).Return(&models.Deployment{}, nil)
	return mockStorage
}
