package pytorch

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	structpb "github.com/golang/protobuf/ptypes/struct"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	kfplugins "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins/kubeflow"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/logs"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	pluginsK8s "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	flytek8sConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	k8sConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	pluginIOMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/kfoperators/common"
	stdlibUtils "github.com/flyteorg/flyte/flytestdlib/utils"
)

const testImage = "image://"
const testImageMaster = "image://master"
const serviceAccount = "pytorch_sa"

var (
	dummyEnvVars = []*core.KeyValuePair{
		{Key: "Env_Var", Value: "Env_Val"},
	}

	testArgs = []string{
		"test-args",
	}

	dummyAnnotations = map[string]string{
		"annotation-key": "annotation-value",
	}
	dummyLabels = map[string]string{
		"label-key": "label-value",
	}

	resourceRequirements = &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:         resource.MustParse("1000m"),
			corev1.ResourceMemory:      resource.MustParse("1Gi"),
			flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:         resource.MustParse("100m"),
			corev1.ResourceMemory:      resource.MustParse("512Mi"),
			flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
		},
	}

	jobName      = "the-job"
	jobNamespace = "pytorch-namespace"
)

func dummyPytorchCustomObj(workers int32) *plugins.DistributedPyTorchTrainingTask {
	return &plugins.DistributedPyTorchTrainingTask{
		Workers: workers,
	}
}

func dummyElasticPytorchCustomObj(workers int32, elasticConfig plugins.ElasticConfig) *plugins.DistributedPyTorchTrainingTask {
	return &plugins.DistributedPyTorchTrainingTask{
		Workers:       workers,
		ElasticConfig: &elasticConfig,
	}
}

func dummyPytorchTaskTemplate(id string, args ...interface{}) *core.TaskTemplate {

	var ptObjJSON string
	var err error

	for _, arg := range args {
		switch t := arg.(type) {
		case *kfplugins.DistributedPyTorchTrainingTask:
			var pytorchCustomObj = t
			ptObjJSON, err = utils.MarshalToString(pytorchCustomObj)
		case *plugins.DistributedPyTorchTrainingTask:
			var pytorchCustomObj = t
			ptObjJSON, err = utils.MarshalToString(pytorchCustomObj)
		default:
			err = fmt.Errorf("Unknown input type %T", t)
		}
	}

	if err != nil {
		panic(err)
	}

	structObj := structpb.Struct{}

	err = stdlibUtils.UnmarshalStringToPb(ptObjJSON, &structObj)
	if err != nil {
		panic(err)
	}

	return &core.TaskTemplate{
		Id:   &core.Identifier{Name: id},
		Type: "container",
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Image: testImage,
				Args:  testArgs,
				Env:   dummyEnvVars,
			},
		},
		Custom: &structObj,
	}
}

func dummyPytorchTaskContext(taskTemplate *core.TaskTemplate, resources *corev1.ResourceRequirements, extendedResources *core.ExtendedResources, containerImage string, pluginState k8s.PluginState) pluginsCore.TaskExecutionContext {
	taskCtx := &mocks.TaskExecutionContext{}
	inputReader := &pluginIOMocks.InputReader{}
	inputReader.EXPECT().GetInputPrefixPath().Return("/input/prefix")
	inputReader.EXPECT().GetInputPath().Return("/input")
	inputReader.EXPECT().Get(mock.Anything).Return(&core.LiteralMap{}, nil)
	taskCtx.EXPECT().InputReader().Return(inputReader)

	outputReader := &pluginIOMocks.OutputWriter{}
	outputReader.EXPECT().GetOutputPath().Return("/data/outputs.pb")
	outputReader.EXPECT().GetOutputPrefixPath().Return("/data/")
	outputReader.EXPECT().GetRawOutputPrefix().Return("")
	outputReader.EXPECT().GetCheckpointPrefix().Return("/checkpoint")
	outputReader.EXPECT().GetPreviousCheckpointsPrefix().Return("/prev")
	taskCtx.EXPECT().OutputWriter().Return(outputReader)

	taskReader := &mocks.TaskReader{}
	taskReader.EXPECT().Read(mock.Anything).Return(taskTemplate, nil)
	taskCtx.EXPECT().TaskReader().Return(taskReader)

	tID := &mocks.TaskExecutionID{}
	tID.EXPECT().GetID().Return(core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})
	tID.EXPECT().GetGeneratedName().Return("some-acceptable-name")
	tID.On("GetUniqueNodeID").Return("an-unique-id")

	overrides := &mocks.TaskOverrides{}
	overrides.EXPECT().GetResources().Return(resources)
	overrides.EXPECT().GetExtendedResources().Return(extendedResources)
	overrides.EXPECT().GetContainerImage().Return(containerImage)
	overrides.EXPECT().GetPodTemplate().Return(nil)

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.EXPECT().GetTaskExecutionID().Return(tID)
	taskExecutionMetadata.EXPECT().GetNamespace().Return("test-namespace")
	taskExecutionMetadata.EXPECT().GetAnnotations().Return(dummyAnnotations)
	taskExecutionMetadata.EXPECT().GetLabels().Return(dummyLabels)
	taskExecutionMetadata.EXPECT().GetOwnerReference().Return(v1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskExecutionMetadata.EXPECT().IsInterruptible().Return(true)
	taskExecutionMetadata.EXPECT().GetOverrides().Return(overrides)
	taskExecutionMetadata.EXPECT().GetK8sServiceAccount().Return(serviceAccount)
	taskExecutionMetadata.EXPECT().GetPlatformResources().Return(&corev1.ResourceRequirements{})
	taskExecutionMetadata.EXPECT().GetEnvironmentVariables().Return(nil)
	taskExecutionMetadata.EXPECT().GetConsoleURL().Return("")
	taskCtx.EXPECT().TaskExecutionMetadata().Return(taskExecutionMetadata)

	pluginStateReaderMock := mocks.PluginStateReader{}
	pluginStateReaderMock.On("Get", mock.AnythingOfType(reflect.TypeOf(&pluginState).String())).Return(
		func(v interface{}) uint8 {
			*(v.(*k8s.PluginState)) = pluginState
			return 0
		},
		func(v interface{}) error {
			return nil
		})

	taskCtx.EXPECT().PluginStateReader().Return(&pluginStateReaderMock)
	return taskCtx
}

func dummyPytorchJobResource(pytorchResourceHandler pytorchOperatorResourceHandler, workers int32, conditionType kubeflowv1.JobConditionType) *kubeflowv1.PyTorchJob {
	var jobConditions []kubeflowv1.JobCondition

	now := time.Now()

	jobCreated := kubeflowv1.JobCondition{
		Type:    kubeflowv1.JobCreated,
		Status:  corev1.ConditionTrue,
		Reason:  "PyTorchJobCreated",
		Message: "PyTorchJob the-job is created.",
		LastUpdateTime: v1.Time{
			Time: now,
		},
		LastTransitionTime: v1.Time{
			Time: now,
		},
	}
	jobRunningActive := kubeflowv1.JobCondition{
		Type:    kubeflowv1.JobRunning,
		Status:  corev1.ConditionTrue,
		Reason:  "PyTorchJobRunning",
		Message: "PyTorchJob the-job is running.",
		LastUpdateTime: v1.Time{
			Time: now.Add(time.Minute),
		},
		LastTransitionTime: v1.Time{
			Time: now.Add(time.Minute),
		},
	}
	jobRunningInactive := *jobRunningActive.DeepCopy()
	jobRunningInactive.Status = corev1.ConditionFalse
	jobSucceeded := kubeflowv1.JobCondition{
		Type:    kubeflowv1.JobSucceeded,
		Status:  corev1.ConditionTrue,
		Reason:  "PyTorchJobSucceeded",
		Message: "PyTorchJob the-job is successfully completed.",
		LastUpdateTime: v1.Time{
			Time: now.Add(2 * time.Minute),
		},
		LastTransitionTime: v1.Time{
			Time: now.Add(2 * time.Minute),
		},
	}
	jobFailed := kubeflowv1.JobCondition{
		Type:    kubeflowv1.JobFailed,
		Status:  corev1.ConditionTrue,
		Reason:  "PyTorchJobFailed",
		Message: "PyTorchJob the-job is failed.",
		LastUpdateTime: v1.Time{
			Time: now.Add(2 * time.Minute),
		},
		LastTransitionTime: v1.Time{
			Time: now.Add(2 * time.Minute),
		},
	}
	jobRestarting := kubeflowv1.JobCondition{
		Type:    kubeflowv1.JobRestarting,
		Status:  corev1.ConditionTrue,
		Reason:  "PyTorchJobRestarting",
		Message: "PyTorchJob the-job is restarting because some replica(s) failed.",
		LastUpdateTime: v1.Time{
			Time: now.Add(3 * time.Minute),
		},
		LastTransitionTime: v1.Time{
			Time: now.Add(3 * time.Minute),
		},
	}

	switch conditionType {
	case kubeflowv1.JobCreated:
		jobConditions = []kubeflowv1.JobCondition{
			jobCreated,
		}
	case kubeflowv1.JobRunning:
		jobConditions = []kubeflowv1.JobCondition{
			jobCreated,
			jobRunningActive,
		}
	case kubeflowv1.JobSucceeded:
		jobConditions = []kubeflowv1.JobCondition{
			jobCreated,
			jobRunningInactive,
			jobSucceeded,
		}
	case kubeflowv1.JobFailed:
		jobConditions = []kubeflowv1.JobCondition{
			jobCreated,
			jobRunningInactive,
			jobFailed,
		}
	case kubeflowv1.JobRestarting:
		jobConditions = []kubeflowv1.JobCondition{
			jobCreated,
			jobRunningInactive,
			jobFailed,
			jobRestarting,
		}
	}

	ptObj := dummyPytorchCustomObj(workers)
	taskTemplate := dummyPytorchTaskTemplate("job1", ptObj)
	resource, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate, resourceRequirements, nil, "", k8s.PluginState{}))
	if err != nil {
		panic(err)
	}

	return &kubeflowv1.PyTorchJob{
		ObjectMeta: v1.ObjectMeta{
			CreationTimestamp: v1.Time{Time: time.Now()},
			Name:              jobName,
			Namespace:         jobNamespace,
		},
		Spec: resource.(*kubeflowv1.PyTorchJob).Spec,
		Status: kubeflowv1.JobStatus{
			Conditions:        jobConditions,
			ReplicaStatuses:   nil,
			StartTime:         nil,
			CompletionTime:    nil,
			LastReconcileTime: nil,
		},
	}
}

func TestBuildResourcePytorchElastic(t *testing.T) {
	pytorchResourceHandler := pytorchOperatorResourceHandler{}

	ptObj := dummyElasticPytorchCustomObj(2, plugins.ElasticConfig{MinReplicas: 1, MaxReplicas: 2, NprocPerNode: 4, RdzvBackend: "c10d"})
	taskTemplate := dummyPytorchTaskTemplate("job2", ptObj)

	resource, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate, resourceRequirements, nil, "", k8s.PluginState{}))
	assert.NoError(t, err)
	assert.NotNil(t, resource)

	pytorchJob, ok := resource.(*kubeflowv1.PyTorchJob)
	assert.True(t, ok)
	assert.Equal(t, int32(2), *pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Replicas)
	assert.NotNil(t, pytorchJob.Spec.ElasticPolicy)
	assert.Equal(t, int32(1), *pytorchJob.Spec.ElasticPolicy.MinReplicas)
	assert.Equal(t, int32(2), *pytorchJob.Spec.ElasticPolicy.MaxReplicas)
	assert.Equal(t, int32(4), *pytorchJob.Spec.ElasticPolicy.NProcPerNode)
	assert.Equal(t, kubeflowv1.RDZVBackend("c10d"), *pytorchJob.Spec.ElasticPolicy.RDZVBackend)

	assert.Equal(t, 1, len(pytorchJob.Spec.PyTorchReplicaSpecs))
	assert.Contains(t, pytorchJob.Spec.PyTorchReplicaSpecs, kubeflowv1.PyTorchJobReplicaTypeWorker)

	for _, replicaSpec := range pytorchJob.Spec.PyTorchReplicaSpecs {
		var hasContainerWithDefaultPytorchName = false
		podSpec := replicaSpec.Template.Spec
		for _, container := range podSpec.Containers {
			if container.Name == kubeflowv1.PyTorchJobDefaultContainerName {
				hasContainerWithDefaultPytorchName = true
			}
		}

		assert.True(t, hasContainerWithDefaultPytorchName)

		// verify TaskExecutionMetadata labels and annotations are copied to the PyTorchJob
		for k, v := range dummyAnnotations {
			assert.Equal(t, v, replicaSpec.Template.ObjectMeta.Annotations[k])
		}
		for k, v := range dummyLabels {
			assert.Equal(t, v, replicaSpec.Template.ObjectMeta.Labels[k])
		}
	}
}

func TestBuildResourcePytorch(t *testing.T) {
	pytorchResourceHandler := pytorchOperatorResourceHandler{}

	ptObj := dummyPytorchCustomObj(100)
	taskTemplate := dummyPytorchTaskTemplate("job3", ptObj)

	res, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate, resourceRequirements, nil, "", k8s.PluginState{}))
	assert.NoError(t, err)
	assert.NotNil(t, res)

	pytorchJob, ok := res.(*kubeflowv1.PyTorchJob)
	assert.True(t, ok)
	assert.Equal(t, int32(100), *pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Replicas)
	assert.Nil(t, pytorchJob.Spec.ElasticPolicy)

	// verify TaskExecutionMetadata labels and annotations are copied to the TensorFlowJob
	for k, v := range dummyAnnotations {
		for _, replicaSpec := range pytorchJob.Spec.PyTorchReplicaSpecs {
			assert.Equal(t, v, replicaSpec.Template.ObjectMeta.Annotations[k])
		}
	}
	for k, v := range dummyLabels {
		for _, replicaSpec := range pytorchJob.Spec.PyTorchReplicaSpecs {
			assert.Equal(t, v, replicaSpec.Template.ObjectMeta.Labels[k])
		}
	}

	for _, replicaSpec := range pytorchJob.Spec.PyTorchReplicaSpecs {
		var hasContainerWithDefaultPytorchName = false
		for _, container := range replicaSpec.Template.Spec.Containers {
			if container.Name == kubeflowv1.PyTorchJobDefaultContainerName {
				hasContainerWithDefaultPytorchName = true
			}

			assert.Equal(t, resourceRequirements.Requests, container.Resources.Requests, fmt.Sprintf(" container.Resources.Requests [%+v]", container.Resources.Requests.Cpu().String()))
			assert.Equal(t, resourceRequirements.Limits, container.Resources.Limits, fmt.Sprintf(" container.Resources.Limits [%+v]", container.Resources.Limits.Cpu().String()))
		}

		assert.True(t, hasContainerWithDefaultPytorchName)
	}
}

func TestBuildResourcePytorchContainerImage(t *testing.T) {
	assert.NoError(t, flytek8sConfig.SetK8sPluginConfig(&flytek8sConfig.K8sPluginConfig{}))

	fixtures := []struct {
		name                   string
		resources              *corev1.ResourceRequirements
		containerImageOverride string
	}{
		{
			"without overrides",
			&corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
				},
			},
			"",
		},
		{
			"with overrides",
			&corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
				},
			},
			"container-image-override",
		},
	}

	testConfigs := []struct {
		name   string
		plugin *plugins.DistributedPyTorchTrainingTask
	}{
		{
			"pytorch",
			dummyPytorchCustomObj(100),
		},
		{
			"elastic pytorch",
			dummyElasticPytorchCustomObj(2, plugins.ElasticConfig{MinReplicas: 1, MaxReplicas: 2, NprocPerNode: 4, RdzvBackend: "c10d"}),
		},
	}

	for _, tCfg := range testConfigs {
		for _, f := range fixtures {
			t.Run(tCfg.name+" "+f.name, func(t *testing.T) {
				taskTemplate := dummyPytorchTaskTemplate("job", tCfg.plugin)
				taskContext := dummyPytorchTaskContext(taskTemplate, f.resources, nil, f.containerImageOverride, k8s.PluginState{})
				pytorchResourceHandler := pytorchOperatorResourceHandler{}
				r, err := pytorchResourceHandler.BuildResource(context.TODO(), taskContext)
				assert.NoError(t, err)
				assert.NotNil(t, r)
				pytorchJob, ok := r.(*kubeflowv1.PyTorchJob)
				assert.True(t, ok)

				for _, replicaSpec := range pytorchJob.Spec.PyTorchReplicaSpecs {
					var expectedContainerImage string
					if len(f.containerImageOverride) > 0 {
						expectedContainerImage = f.containerImageOverride
					} else {
						expectedContainerImage = testImage
					}
					assert.Equal(t, expectedContainerImage, replicaSpec.Template.Spec.Containers[0].Image)
				}
			})
		}
	}
}

func TestBuildResourcePytorchExtendedResources(t *testing.T) {
	assert.NoError(t, flytek8sConfig.SetK8sPluginConfig(&flytek8sConfig.K8sPluginConfig{
		GpuDeviceNodeLabel:                 "gpu-node-label",
		GpuPartitionSizeNodeLabel:          "gpu-partition-size",
		GpuResourceName:                    flytek8s.ResourceNvidiaGPU,
		AddTolerationsForExtendedResources: []string{"nvidia.com/gpu"},
	}))

	fixtures := []struct {
		name                      string
		resources                 *corev1.ResourceRequirements
		extendedResourcesBase     *core.ExtendedResources
		extendedResourcesOverride *core.ExtendedResources
		expectedNsr               []corev1.NodeSelectorTerm
		expectedTol               []corev1.Toleration
	}{
		{
			"without overrides",
			&corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
				},
			},
			&core.ExtendedResources{
				GpuAccelerator: &core.GPUAccelerator{
					Device: "nvidia-tesla-t4",
				},
			},
			nil,
			[]corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						corev1.NodeSelectorRequirement{
							Key:      "gpu-node-label",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"nvidia-tesla-t4"},
						},
					},
				},
			},
			[]corev1.Toleration{
				{
					Key:      "gpu-node-label",
					Value:    "nvidia-tesla-t4",
					Operator: corev1.TolerationOpEqual,
					Effect:   corev1.TaintEffectNoSchedule,
				},
				{
					Key:      "nvidia.com/gpu",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
		},
		{
			"with overrides",
			&corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
				},
			},
			&core.ExtendedResources{
				GpuAccelerator: &core.GPUAccelerator{
					Device: "nvidia-tesla-t4",
				},
			},
			&core.ExtendedResources{
				GpuAccelerator: &core.GPUAccelerator{
					Device: "nvidia-tesla-a100",
					PartitionSizeValue: &core.GPUAccelerator_PartitionSize{
						PartitionSize: "1g.5gb",
					},
				},
			},
			[]corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						corev1.NodeSelectorRequirement{
							Key:      "gpu-node-label",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"nvidia-tesla-a100"},
						},
						corev1.NodeSelectorRequirement{
							Key:      "gpu-partition-size",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"1g.5gb"},
						},
					},
				},
			},
			[]corev1.Toleration{
				{
					Key:      "gpu-node-label",
					Value:    "nvidia-tesla-a100",
					Operator: corev1.TolerationOpEqual,
					Effect:   corev1.TaintEffectNoSchedule,
				},
				{
					Key:      "gpu-partition-size",
					Value:    "1g.5gb",
					Operator: corev1.TolerationOpEqual,
					Effect:   corev1.TaintEffectNoSchedule,
				},
				{
					Key:      "nvidia.com/gpu",
					Operator: corev1.TolerationOpExists,
					Effect:   corev1.TaintEffectNoSchedule,
				},
			},
		},
	}

	testConfigs := []struct {
		name   string
		plugin *plugins.DistributedPyTorchTrainingTask
	}{
		{
			"pytorch",
			dummyPytorchCustomObj(100),
		},
		{
			"elastic pytorch",
			dummyElasticPytorchCustomObj(2, plugins.ElasticConfig{MinReplicas: 1, MaxReplicas: 2, NprocPerNode: 4, RdzvBackend: "c10d"}),
		},
	}

	for _, tCfg := range testConfigs {
		for _, f := range fixtures {
			t.Run(tCfg.name+" "+f.name, func(t *testing.T) {
				taskTemplate := dummyPytorchTaskTemplate("job", tCfg.plugin)
				taskTemplate.ExtendedResources = f.extendedResourcesBase
				taskContext := dummyPytorchTaskContext(taskTemplate, f.resources, f.extendedResourcesOverride, "", k8s.PluginState{})
				pytorchResourceHandler := pytorchOperatorResourceHandler{}
				r, err := pytorchResourceHandler.BuildResource(context.TODO(), taskContext)
				assert.NoError(t, err)
				assert.NotNil(t, r)
				pytorchJob, ok := r.(*kubeflowv1.PyTorchJob)
				assert.True(t, ok)

				for _, replicaSpec := range pytorchJob.Spec.PyTorchReplicaSpecs {
					assert.EqualValues(
						t,
						f.expectedNsr,
						replicaSpec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
					)
					assert.EqualValues(
						t,
						f.expectedTol,
						replicaSpec.Template.Spec.Tolerations,
					)
				}
			})
		}
	}
}

func TestGetTaskPhase(t *testing.T) {
	pytorchResourceHandler := pytorchOperatorResourceHandler{}
	ctx := context.TODO()

	dummyPytorchJobResourceCreator := func(conditionType kubeflowv1.JobConditionType) *kubeflowv1.PyTorchJob {
		return dummyPytorchJobResource(pytorchResourceHandler, 2, conditionType)
	}

	taskCtx := dummyPytorchTaskContext(dummyPytorchTaskTemplate("", dummyPytorchCustomObj(2)), resourceRequirements, nil, "", k8s.PluginState{})
	taskPhase, err := pytorchResourceHandler.GetTaskPhase(ctx, taskCtx, dummyPytorchJobResourceCreator(kubeflowv1.JobCreated))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = pytorchResourceHandler.GetTaskPhase(ctx, taskCtx, dummyPytorchJobResourceCreator(kubeflowv1.JobRunning))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = pytorchResourceHandler.GetTaskPhase(ctx, taskCtx, dummyPytorchJobResourceCreator(kubeflowv1.JobSucceeded))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = pytorchResourceHandler.GetTaskPhase(ctx, taskCtx, dummyPytorchJobResourceCreator(kubeflowv1.JobFailed))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	taskPhase, err = pytorchResourceHandler.GetTaskPhase(ctx, taskCtx, dummyPytorchJobResourceCreator(kubeflowv1.JobRestarting))
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	// Training operator did not modify the job even though it is not suspended
	pytorchJob := dummyPytorchJobResourceCreator(kubeflowv1.JobCreated)
	pytorchJob.CreationTimestamp = v1.Time{Time: time.Now().Add(-time.Hour)}
	pytorchJob.Status.StartTime = nil
	taskPhase, err = pytorchResourceHandler.GetTaskPhase(ctx, taskCtx, pytorchJob)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "kubeflow operator hasn't updated")
	assert.Equal(t, pluginsCore.PhaseInfoUndefined, taskPhase)

	// Training operator did not modify the job because it is suspended
	pytorchJobSuspended := dummyPytorchJobResourceCreator(kubeflowv1.JobCreated)
	pytorchJobSuspended.CreationTimestamp = v1.Time{Time: time.Now().Add(-time.Hour)}
	pytorchJobSuspended.Status.StartTime = nil
	suspend := true
	pytorchJobSuspended.Spec.RunPolicy.Suspend = &suspend
	taskPhase, err = pytorchResourceHandler.GetTaskPhase(ctx, taskCtx, pytorchJobSuspended)
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
}

func TestGetTaskPhaseIncreasePhaseVersion(t *testing.T) {
	pytorchResourceHandler := pytorchOperatorResourceHandler{}
	ctx := context.TODO()

	pluginState := k8s.PluginState{
		Phase:        pluginsCore.PhaseQueued,
		PhaseVersion: pluginsCore.DefaultPhaseVersion,
		Reason:       "task submitted to K8s",
	}
	taskCtx := dummyPytorchTaskContext(dummyPytorchTaskTemplate("", dummyPytorchCustomObj(2)), resourceRequirements, nil, "", pluginState)

	taskPhase, err := pytorchResourceHandler.GetTaskPhase(ctx, taskCtx, dummyPytorchJobResource(pytorchResourceHandler, 4, kubeflowv1.JobCreated))

	assert.NoError(t, err)
	assert.Equal(t, taskPhase.Version(), pluginsCore.DefaultPhaseVersion+1)
}

func TestGetLogs(t *testing.T) {
	assert.NoError(t, logs.SetLogConfig(&logs.LogConfig{
		IsKubernetesEnabled: true,
		KubernetesURL:       "k8s.com",
	}))

	hasMaster := true
	workers := int32(2)

	pytorchResourceHandler := pytorchOperatorResourceHandler{}
	pytorchJob := dummyPytorchJobResource(pytorchResourceHandler, workers, kubeflowv1.JobRunning)
	taskTemplate := dummyPytorchTaskTemplate("", dummyPytorchCustomObj(workers))
	taskCtx := dummyPytorchTaskContext(taskTemplate, resourceRequirements, nil, "", k8s.PluginState{})
	jobLogs, err := common.GetLogs(taskCtx, common.PytorchTaskType, pytorchJob.ObjectMeta, taskTemplate, hasMaster, workers, 0, 0, 0, kubeflowv1.PyTorchJobDefaultContainerName)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-master-0/pod?namespace=pytorch-namespace", jobNamespace, jobName), jobLogs[0].GetUri())
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=pytorch-namespace", jobNamespace, jobName), jobLogs[1].GetUri())
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-1/pod?namespace=pytorch-namespace", jobNamespace, jobName), jobLogs[2].GetUri())
}

func TestGetLogsElastic(t *testing.T) {
	assert.NoError(t, logs.SetLogConfig(&logs.LogConfig{
		IsKubernetesEnabled: true,
		KubernetesURL:       "k8s.com",
	}))

	hasMaster := false
	workers := int32(2)

	pytorchResourceHandler := pytorchOperatorResourceHandler{}
	pytorchJob := dummyPytorchJobResource(pytorchResourceHandler, workers, kubeflowv1.JobRunning)
	taskTemplate := dummyPytorchTaskTemplate("", dummyPytorchCustomObj(workers))
	taskCtx := dummyPytorchTaskContext(taskTemplate, resourceRequirements, nil, "", k8s.PluginState{})
	jobLogs, err := common.GetLogs(taskCtx, common.PytorchTaskType, pytorchJob.ObjectMeta, taskTemplate, hasMaster, workers, 0, 0, 0, kubeflowv1.PyTorchJobDefaultContainerName)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=pytorch-namespace", jobNamespace, jobName), jobLogs[0].GetUri())
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-1/pod?namespace=pytorch-namespace", jobNamespace, jobName), jobLogs[1].GetUri())
}

func TestGetProperties(t *testing.T) {
	config := k8sConfig.GetK8sPluginConfig()
	pytorchResourceHandler := pytorchOperatorResourceHandler{}

	expected := k8s.PluginProperties{}
	assert.Equal(t, expected, pytorchResourceHandler.GetProperties())

	config.EnableDistributedErrorAggregation = true
	expected = k8s.PluginProperties{
		ErrorAggregationStrategy: k8s.EarliestErrorAggregationStrategy,
	}
	assert.Equal(t, expected, pytorchResourceHandler.GetProperties())
}

func TestReplicaCounts(t *testing.T) {
	for _, test := range []struct {
		name               string
		workerReplicaCount int32
		expectError        bool
		contains           []kubeflowv1.ReplicaType
		notContains        []kubeflowv1.ReplicaType
	}{
		{"NoWorkers", 0, false, []kubeflowv1.ReplicaType{kubeflowv1.PyTorchJobReplicaTypeMaster}, []kubeflowv1.ReplicaType{}},
		{"Works", 1, false, []kubeflowv1.ReplicaType{kubeflowv1.PyTorchJobReplicaTypeMaster, kubeflowv1.PyTorchJobReplicaTypeWorker}, []kubeflowv1.ReplicaType{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			pytorchResourceHandler := pytorchOperatorResourceHandler{}

			ptObj := dummyPytorchCustomObj(test.workerReplicaCount)
			taskTemplate := dummyPytorchTaskTemplate("the job", ptObj)

			res, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate, resourceRequirements, nil, "", k8s.PluginState{}))
			if test.expectError {
				assert.Error(t, err)
				assert.Nil(t, res)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)

			job, ok := res.(*kubeflowv1.PyTorchJob)
			assert.True(t, ok)

			assert.Len(t, job.Spec.PyTorchReplicaSpecs, len(test.contains))
			for _, replicaType := range test.contains {
				assert.Contains(t, job.Spec.PyTorchReplicaSpecs, replicaType)
			}
			for _, replicaType := range test.notContains {
				assert.NotContains(t, job.Spec.PyTorchReplicaSpecs, replicaType)
			}
		})
	}
}

func TestBuildResourcePytorchV1(t *testing.T) {
	taskConfigs := []*kfplugins.DistributedPyTorchTrainingTask{
		{
			MasterReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Image: testImageMaster,
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "250m"},
						{Name: core.Resources_MEMORY, Value: "250Mi"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "500m"},
						{Name: core.Resources_MEMORY, Value: "500Mi"},
					},
				},
				RestartPolicy: kfplugins.RestartPolicy_RESTART_POLICY_ALWAYS,
			},
			WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Replicas: 100,
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "1024m"},
						{Name: core.Resources_MEMORY, Value: "1Gi"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "2048m"},
						{Name: core.Resources_MEMORY, Value: "2Gi"},
					},
				},
			},
		},
		{
			MasterReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Common: &kfplugins.CommonReplicaSpec{
					Image: testImageMaster,
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "250m"},
							{Name: core.Resources_MEMORY, Value: "250Mi"},
						},
						Limits: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "500m"},
							{Name: core.Resources_MEMORY, Value: "500Mi"},
						},
					},
					RestartPolicy: kfplugins.RestartPolicy_RESTART_POLICY_ALWAYS,
				},
			},
			WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Common: &kfplugins.CommonReplicaSpec{
					Replicas: 100,
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "1024m"},
							{Name: core.Resources_MEMORY, Value: "1Gi"},
						},
						Limits: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "2048m"},
							{Name: core.Resources_MEMORY, Value: "2Gi"},
						},
					},
				},
			},
		},
	}

	for _, taskConfig := range taskConfigs {
		masterResourceRequirements := &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("250m"),
				corev1.ResourceMemory: resource.MustParse("250Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("500m"),
				corev1.ResourceMemory: resource.MustParse("500Mi"),
			},
		}

		workerResourceRequirements := &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1024m"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2048m"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
		}

		config := k8sConfig.GetK8sPluginConfig()
		config.EnableDistributedErrorAggregation = true
		pytorchResourceHandler := pytorchOperatorResourceHandler{}

		taskTemplate := dummyPytorchTaskTemplate("job4", taskConfig)
		taskTemplate.TaskTypeVersion = 1

		res, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate, resourceRequirements, nil, "", k8s.PluginState{}))
		assert.NoError(t, err)
		assert.NotNil(t, res)

		pytorchJob, ok := res.(*kubeflowv1.PyTorchJob)
		assert.True(t, ok)

		assert.Equal(t, int32(100), *pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Replicas)
		assert.Equal(t, int32(1), *pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Replicas)

		assert.Equal(t, testImageMaster, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.Spec.Containers[0].Image)
		assert.Equal(t, testImage, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.Spec.Containers[0].Image)

		assert.Equal(t, *masterResourceRequirements, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.Spec.Containers[0].Resources)
		assert.Equal(t, *workerResourceRequirements, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.Spec.Containers[0].Resources)

		assert.Equal(t, kubeflowv1.RestartPolicyAlways, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].RestartPolicy)
		assert.Equal(t, kubeflowv1.RestartPolicyNever, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].RestartPolicy)

		assert.Nil(t, pytorchJob.Spec.RunPolicy.CleanPodPolicy)
		assert.Nil(t, pytorchJob.Spec.RunPolicy.BackoffLimit)
		assert.Nil(t, pytorchJob.Spec.RunPolicy.TTLSecondsAfterFinished)
		assert.Nil(t, pytorchJob.Spec.RunPolicy.ActiveDeadlineSeconds)

		assert.Nil(t, pytorchJob.Spec.ElasticPolicy)

		// validate plugin specific environment variables
		workerContainerEnv := pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.Spec.Containers[0].Env
		assert.Equal(t,
			[]apiv1.EnvVar{
				{
					Name: pluginsK8s.FlyteInternalWorkerNameEnvVarKey,
					ValueFrom: &apiv1.EnvVarSource{
						FieldRef: &apiv1.ObjectFieldSelector{
							FieldPath: "metadata.name",
						},
					},
				},
				{
					Name:  pluginsK8s.FlyteInternalDistErrorStrategyEnvVarKey,
					Value: "Earliest",
				},
			},
			workerContainerEnv[len(workerContainerEnv)-2:],
		)
	}
}

func TestBuildResourcePytorchV1WithRunPolicy(t *testing.T) {
	taskConfigs := []*kfplugins.DistributedPyTorchTrainingTask{
		{
			WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Replicas: 100,
			},
			RunPolicy: &kfplugins.RunPolicy{
				CleanPodPolicy:          kfplugins.CleanPodPolicy_CLEANPOD_POLICY_ALL,
				BackoffLimit:            100,
				ActiveDeadlineSeconds:   1000,
				TtlSecondsAfterFinished: 10000,
			},
		},
		{
			WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Common: &kfplugins.CommonReplicaSpec{
					Replicas: 100,
				},
			},
			RunPolicy: &kfplugins.RunPolicy{
				CleanPodPolicy:          kfplugins.CleanPodPolicy_CLEANPOD_POLICY_ALL,
				BackoffLimit:            100,
				ActiveDeadlineSeconds:   1000,
				TtlSecondsAfterFinished: 10000,
			},
		},
	}

	for _, taskConfig := range taskConfigs {
		pytorchResourceHandler := pytorchOperatorResourceHandler{}

		taskTemplate := dummyPytorchTaskTemplate("job5", taskConfig)
		taskTemplate.TaskTypeVersion = 1

		res, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate, resourceRequirements, nil, "", k8s.PluginState{}))
		assert.NoError(t, err)
		assert.NotNil(t, res)

		pytorchJob, ok := res.(*kubeflowv1.PyTorchJob)
		assert.True(t, ok)
		assert.Equal(t, int32(100), *pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Replicas)
		assert.Equal(t, int32(1), *pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Replicas)
		assert.Equal(t, kubeflowv1.CleanPodPolicyAll, *pytorchJob.Spec.RunPolicy.CleanPodPolicy)
		assert.Equal(t, int32(100), *pytorchJob.Spec.RunPolicy.BackoffLimit)
		assert.Equal(t, int64(1000), *pytorchJob.Spec.RunPolicy.ActiveDeadlineSeconds)
		assert.Equal(t, int32(10000), *pytorchJob.Spec.RunPolicy.TTLSecondsAfterFinished)
	}
}

func TestBuildResourcePytorchV1WithOnlyWorkerSpec(t *testing.T) {
	taskConfigs := []*kfplugins.DistributedPyTorchTrainingTask{
		{
			WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Replicas: 100,
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "1024m"},
						{Name: core.Resources_MEMORY, Value: "1Gi"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "2048m"},
						{Name: core.Resources_MEMORY, Value: "2Gi"},
					},
				},
			},
		},
		{
			WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Common: &kfplugins.CommonReplicaSpec{
					Replicas: 100,
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "1024m"},
							{Name: core.Resources_MEMORY, Value: "1Gi"},
						},
						Limits: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "2048m"},
							{Name: core.Resources_MEMORY, Value: "2Gi"},
						},
					},
				},
			},
		},
	}

	for _, taskConfig := range taskConfigs {
		// Master Replica should use resource from task override if not set
		taskOverrideResourceRequirements := &corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:         resource.MustParse("1000m"),
				corev1.ResourceMemory:      resource.MustParse("1Gi"),
				flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:         resource.MustParse("100m"),
				corev1.ResourceMemory:      resource.MustParse("512Mi"),
				flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
			},
		}

		workerResourceRequirements := &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1024m"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("2048m"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
		}

		pytorchResourceHandler := pytorchOperatorResourceHandler{}

		taskTemplate := dummyPytorchTaskTemplate("job5", taskConfig)
		taskTemplate.TaskTypeVersion = 1

		res, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate, resourceRequirements, nil, "", k8s.PluginState{}))
		assert.NoError(t, err)
		assert.NotNil(t, res)

		pytorchJob, ok := res.(*kubeflowv1.PyTorchJob)
		assert.True(t, ok)

		assert.Equal(t, int32(100), *pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Replicas)
		assert.Equal(t, int32(1), *pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Replicas)

		assert.Equal(t, testImage, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.Spec.Containers[0].Image)
		assert.Equal(t, testImage, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.Spec.Containers[0].Image)

		assert.Equal(t, *taskOverrideResourceRequirements, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.Spec.Containers[0].Resources)
		assert.Equal(t, *workerResourceRequirements, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.Spec.Containers[0].Resources)

		assert.Equal(t, kubeflowv1.RestartPolicyNever, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].RestartPolicy)
		assert.Equal(t, kubeflowv1.RestartPolicyNever, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].RestartPolicy)

		assert.Nil(t, pytorchJob.Spec.ElasticPolicy)
	}
}

func TestBuildResourcePytorchV1ResourceTolerations(t *testing.T) {
	gpuToleration := corev1.Toleration{
		Key:      "nvidia.com/gpu",
		Value:    "present",
		Operator: corev1.TolerationOpEqual,
		Effect:   corev1.TaintEffectNoSchedule,
	}
	assert.NoError(t, flytek8sConfig.SetK8sPluginConfig(&flytek8sConfig.K8sPluginConfig{
		GpuResourceName: flytek8s.ResourceNvidiaGPU,
		ResourceTolerations: map[corev1.ResourceName][]corev1.Toleration{
			flytek8s.ResourceNvidiaGPU: {gpuToleration},
		},
	}))

	taskConfigs := []*kfplugins.DistributedPyTorchTrainingTask{
		{
			MasterReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "250m"},
						{Name: core.Resources_MEMORY, Value: "250Mi"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "500m"},
						{Name: core.Resources_MEMORY, Value: "500Mi"},
					},
				},
			},
			WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Replicas: 100,
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "1024m"},
						{Name: core.Resources_MEMORY, Value: "1Gi"},
						{Name: core.Resources_GPU, Value: "1"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "2048m"},
						{Name: core.Resources_MEMORY, Value: "2Gi"},
						{Name: core.Resources_GPU, Value: "1"},
					},
				},
			},
		},
		{
			MasterReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Common: &kfplugins.CommonReplicaSpec{
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "250m"},
							{Name: core.Resources_MEMORY, Value: "250Mi"},
						},
						Limits: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "500m"},
							{Name: core.Resources_MEMORY, Value: "500Mi"},
						},
					},
				},
			},
			WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Common: &kfplugins.CommonReplicaSpec{
					Replicas: 100,
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "1024m"},
							{Name: core.Resources_MEMORY, Value: "1Gi"},
							{Name: core.Resources_GPU, Value: "1"},
						},
						Limits: []*core.Resources_ResourceEntry{
							{Name: core.Resources_CPU, Value: "2048m"},
							{Name: core.Resources_MEMORY, Value: "2Gi"},
							{Name: core.Resources_GPU, Value: "1"},
						},
					},
				},
			},
		},
	}

	for _, taskConfig := range taskConfigs {
		pytorchResourceHandler := pytorchOperatorResourceHandler{}

		taskTemplate := dummyPytorchTaskTemplate("job4", taskConfig)
		taskTemplate.TaskTypeVersion = 1

		res, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate, resourceRequirements, nil, "", k8s.PluginState{}))
		assert.NoError(t, err)
		assert.NotNil(t, res)

		pytorchJob, ok := res.(*kubeflowv1.PyTorchJob)
		assert.True(t, ok)

		assert.NotContains(t, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.Spec.Tolerations, gpuToleration)
		assert.Contains(t, pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.Spec.Tolerations, gpuToleration)
	}
}

func TestBuildResourcePytorchV1WithElastic(t *testing.T) {
	taskConfigs := []*kfplugins.DistributedPyTorchTrainingTask{
		{
			WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Replicas: 2,
			},
			ElasticConfig: &kfplugins.ElasticConfig{MinReplicas: 1, MaxReplicas: 2, NprocPerNode: 4, RdzvBackend: "c10d"},
		},
		{
			WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Common: &kfplugins.CommonReplicaSpec{
					Replicas: 2,
				},
			},
			ElasticConfig: &kfplugins.ElasticConfig{MinReplicas: 1, MaxReplicas: 2, NprocPerNode: 4, RdzvBackend: "c10d"},
		},
	}

	for _, taskConfig := range taskConfigs {
		taskTemplate := dummyPytorchTaskTemplate("job5", taskConfig)
		taskTemplate.TaskTypeVersion = 1

		pytorchResourceHandler := pytorchOperatorResourceHandler{}
		resource, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate, resourceRequirements, nil, "", k8s.PluginState{}))
		assert.NoError(t, err)
		assert.NotNil(t, resource)

		pytorchJob, ok := resource.(*kubeflowv1.PyTorchJob)
		assert.True(t, ok)
		assert.Equal(t, int32(2), *pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Replicas)
		assert.NotNil(t, pytorchJob.Spec.ElasticPolicy)
		assert.Equal(t, int32(1), *pytorchJob.Spec.ElasticPolicy.MinReplicas)
		assert.Equal(t, int32(2), *pytorchJob.Spec.ElasticPolicy.MaxReplicas)
		assert.Equal(t, int32(4), *pytorchJob.Spec.ElasticPolicy.NProcPerNode)
		assert.Equal(t, kubeflowv1.RDZVBackend("c10d"), *pytorchJob.Spec.ElasticPolicy.RDZVBackend)

		assert.Equal(t, 1, len(pytorchJob.Spec.PyTorchReplicaSpecs))
		assert.Contains(t, pytorchJob.Spec.PyTorchReplicaSpecs, kubeflowv1.PyTorchJobReplicaTypeWorker)

		var hasContainerWithDefaultPytorchName = false

		for _, container := range pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.Spec.Containers {
			if container.Name == kubeflowv1.PyTorchJobDefaultContainerName {
				hasContainerWithDefaultPytorchName = true
			}
		}

		assert.True(t, hasContainerWithDefaultPytorchName)
	}
}

func TestBuildResourcePytorchV1WithDifferentWorkersNumber(t *testing.T) {
	taskConfigs := []*kfplugins.DistributedPyTorchTrainingTask{
		{
			// Test case 1: Zero workers - should only have master
			WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Replicas: 0,
			},
			MasterReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Image: testImageMaster,
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "250m"},
						{Name: core.Resources_MEMORY, Value: "250Mi"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "500m"},
						{Name: core.Resources_MEMORY, Value: "500Mi"},
					},
				},
			},
		},
		{
			// Test case 2: One worker - should have both master and worker
			WorkerReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Replicas: 1,
			},
			MasterReplicas: &kfplugins.DistributedPyTorchTrainingReplicaSpec{
				Image: testImageMaster,
				Resources: &core.Resources{
					Requests: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "250m"},
						{Name: core.Resources_MEMORY, Value: "250Mi"},
					},
					Limits: []*core.Resources_ResourceEntry{
						{Name: core.Resources_CPU, Value: "500m"},
						{Name: core.Resources_MEMORY, Value: "500Mi"},
					},
				},
			},
		},
	}

	for i, taskConfig := range taskConfigs {
		t.Run(fmt.Sprintf("Case %d", i+1), func(t *testing.T) {
			pytorchResourceHandler := pytorchOperatorResourceHandler{}

			taskTemplate := dummyPytorchTaskTemplate("job5", taskConfig)
			taskTemplate.TaskTypeVersion = 1

			res, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate, resourceRequirements, nil, "", k8s.PluginState{}))
			assert.NoError(t, err)
			assert.NotNil(t, res)

			pytorchJob, ok := res.(*kubeflowv1.PyTorchJob)
			assert.True(t, ok)

			if taskConfig.GetWorkerReplicas().GetReplicas() == 0 {
				// Should only contain master spec
				assert.Equal(t, 1, len(pytorchJob.Spec.PyTorchReplicaSpecs))
				assert.Contains(t, pytorchJob.Spec.PyTorchReplicaSpecs, kubeflowv1.PyTorchJobReplicaTypeMaster)
				assert.NotContains(t, pytorchJob.Spec.PyTorchReplicaSpecs, kubeflowv1.PyTorchJobReplicaTypeWorker)

				// Verify master spec details
				masterSpec := pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster]
				assert.Equal(t, int32(1), *masterSpec.Replicas)
				assert.Equal(t, testImageMaster, masterSpec.Template.Spec.Containers[0].Image)
			} else {
				// Should contain both master and worker specs
				assert.Equal(t, 2, len(pytorchJob.Spec.PyTorchReplicaSpecs))
				assert.Contains(t, pytorchJob.Spec.PyTorchReplicaSpecs, kubeflowv1.PyTorchJobReplicaTypeMaster)
				assert.Contains(t, pytorchJob.Spec.PyTorchReplicaSpecs, kubeflowv1.PyTorchJobReplicaTypeWorker)

				// Verify master spec details
				masterSpec := pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster]
				assert.Equal(t, int32(1), *masterSpec.Replicas)
				assert.Equal(t, testImageMaster, masterSpec.Template.Spec.Containers[0].Image)

				// Verify worker spec details
				workerSpec := pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker]
				assert.Equal(t, int32(1), *workerSpec.Replicas)
			}
		})
	}
}

func TestParseElasticConfig(t *testing.T) {
	elasticConfig := plugins.ElasticConfig{MinReplicas: 1, MaxReplicas: 2, NprocPerNode: 4, RdzvBackend: "c10d"}
	elasticPolicy := ParseElasticConfig(&elasticConfig)
	assert.Equal(t, int32(1), *elasticPolicy.MinReplicas)
	assert.Equal(t, int32(2), *elasticPolicy.MaxReplicas)
	assert.Equal(t, int32(4), *elasticPolicy.NProcPerNode)
	assert.Equal(t, kubeflowv1.RDZVBackend("c10d"), *elasticPolicy.RDZVBackend)
}

func TestGetReplicaCount(t *testing.T) {
	pytorchResourceHandler := pytorchOperatorResourceHandler{}
	tfObj := dummyPytorchCustomObj(1)
	taskTemplate := dummyPytorchTaskTemplate("the job", tfObj)
	resource, err := pytorchResourceHandler.BuildResource(context.TODO(), dummyPytorchTaskContext(taskTemplate, resourceRequirements, nil, "", k8s.PluginState{}))
	assert.NoError(t, err)
	assert.NotNil(t, resource)
	PytorchJob, ok := resource.(*kubeflowv1.PyTorchJob)
	assert.True(t, ok)

	assert.NotNil(t, common.GetReplicaCount(PytorchJob.Spec.PyTorchReplicaSpecs, kubeflowv1.PyTorchJobReplicaTypeWorker))
}
