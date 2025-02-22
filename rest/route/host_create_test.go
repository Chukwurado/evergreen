package route

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/evergreen-ci/birch"
	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/apimodels"
	"github.com/evergreen-ci/evergreen/cloud"
	"github.com/evergreen-ci/evergreen/db"
	"github.com/evergreen-ci/evergreen/model/distro"
	"github.com/evergreen-ci/evergreen/model/host"
	"github.com/evergreen-ci/evergreen/model/task"
	"github.com/evergreen-ci/evergreen/rest/data"
	"github.com/evergreen-ci/gimlet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMakeIntentHost(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	require.NoError(db.ClearCollections(distro.Collection, host.Collection, task.Collection))
	handler := hostCreateHandler{
		sc: &data.DBConnector{},
	}

	d := distro.Distro{
		Id:       "archlinux-test",
		Aliases:  []string{"archlinux-alias"},
		Provider: evergreen.ProviderNameEc2OnDemand,
		ProviderSettingsList: []*birch.Document{birch.NewDocument(
			birch.EC.String("ami", "ami-123456"),
			birch.EC.String("region", "us-east-1"),
			birch.EC.String("instance_type", "t1.micro"),
			birch.EC.String("subnet_id", "subnet-12345678"),
			birch.EC.SliceString("security_group_ids", []string{"abcdef"}),
		)},
	}
	require.NoError(d.Insert())

	sampleTask := &task.Task{
		Id: "task-id",
	}
	require.NoError(sampleTask.Insert())

	// spawn an evergreen distro
	c := apimodels.CreateHost{
		Distro:              "archlinux-test",
		CloudProvider:       "ec2",
		NumHosts:            "1",
		Scope:               "task",
		SetupTimeoutSecs:    600,
		TeardownTimeoutSecs: 21600,
		KeyName:             "mock_key",
	}
	handler.createHost = c
	handler.taskID = "task-id"
	h, err := handler.sc.MakeIntentHost(handler.taskID, "", "", handler.createHost)
	assert.NoError(err)
	require.NotNil(h)

	assert.Equal("archlinux-test", h.Distro.Id)
	assert.Equal(evergreen.ProviderNameEc2OnDemand, h.Provider)
	assert.Equal(evergreen.ProviderNameEc2OnDemand, h.Distro.Provider)
	assert.Equal(distro.BootstrapMethodNone, h.Distro.BootstrapSettings.Method, "host provisioning should be set to none by default")

	assert.Equal("task-id", h.SpawnOptions.TaskID)
	ec2Settings := &cloud.EC2ProviderSettings{}
	require.Len(h.Distro.ProviderSettingsList, 1)
	assert.NoError(ec2Settings.FromDistroSettings(h.Distro, ""))
	assert.Equal("ami-123456", ec2Settings.AMI)
	assert.Equal("mock_key", ec2Settings.KeyName)
	assert.Equal(true, ec2Settings.IsVpc)

	// test roundtripping
	h, err = host.FindOneByIdOrTag(h.Id)
	assert.NoError(err)
	require.NotNil(h)
	ec2Settings2 := &cloud.EC2ProviderSettings{}
	require.Len(h.Distro.ProviderSettingsList, 1)
	assert.NoError(ec2Settings2.FromDistroSettings(h.Distro, ""))
	assert.Equal("ami-123456", ec2Settings2.AMI)
	assert.Equal("mock_key", ec2Settings2.KeyName)
	assert.Equal(true, ec2Settings2.IsVpc)

	// scope to build
	require.NoError(db.ClearCollections(task.Collection))
	myTask := task.Task{
		Id:      "task-id",
		BuildId: "build-id",
	}
	require.NoError(myTask.Insert())
	c = apimodels.CreateHost{
		Distro:              "archlinux-test",
		CloudProvider:       "ec2",
		NumHosts:            "1",
		Scope:               "build",
		SetupTimeoutSecs:    600,
		TeardownTimeoutSecs: 21600,
		KeyName:             "mock_key",
	}
	handler.createHost = c
	handler.taskID = "task-id"
	h, err = handler.sc.MakeIntentHost(handler.taskID, "", "", handler.createHost)
	assert.NoError(err)
	assert.NotNil(h)
	ec2Settings = &cloud.EC2ProviderSettings{}
	assert.NoError(ec2Settings.FromDistroSettings(h.Distro, ""))
	assert.Equal("build-id", h.SpawnOptions.BuildID)
	assert.Equal("mock_key", ec2Settings.KeyName)
	assert.Equal(true, ec2Settings.IsVpc)
	assert.Equal(distro.BootstrapMethodNone, h.Distro.BootstrapSettings.Method, "host provisioning should be set to none by default")

	// Using an alias should resolve to the actual distro
	c = apimodels.CreateHost{
		Distro:              "archlinux-alias",
		CloudProvider:       "ec2",
		NumHosts:            "1",
		Scope:               "task",
		SetupTimeoutSecs:    600,
		TeardownTimeoutSecs: 21600,
		KeyName:             "mock_key",
	}
	handler.createHost = c
	handler.taskID = "task-id"
	h, err = handler.sc.MakeIntentHost(handler.taskID, "", "", handler.createHost)
	require.NoError(err)
	require.NotNil(h)

	assert.Equal("archlinux-test", h.Distro.Id)
	assert.Equal(evergreen.ProviderNameEc2OnDemand, h.Provider)
	assert.Equal(evergreen.ProviderNameEc2OnDemand, h.Distro.Provider)
	assert.Equal(distro.BootstrapMethodNone, h.Distro.BootstrapSettings.Method, "host provisioning should be set to none by default")

	assert.Equal("task-id", h.SpawnOptions.TaskID)
	ec2Settings = &cloud.EC2ProviderSettings{}
	require.Len(h.Distro.ProviderSettingsList, 1)
	assert.NoError(ec2Settings.FromDistroSettings(h.Distro, ""))
	assert.Equal("ami-123456", ec2Settings.AMI)
	assert.Equal("mock_key", ec2Settings.KeyName)
	assert.Equal(true, ec2Settings.IsVpc)

	// spawn a spot evergreen distro
	c = apimodels.CreateHost{
		Distro:              "archlinux-test",
		CloudProvider:       "ec2",
		NumHosts:            "1",
		Scope:               "task",
		SetupTimeoutSecs:    600,
		TeardownTimeoutSecs: 21600,
		Spot:                true,
		KeyName:             "mock_key",
	}
	handler.createHost = c
	h, err = handler.sc.MakeIntentHost(handler.taskID, "", "", handler.createHost)
	assert.NoError(err)
	assert.NotNil(h)
	assert.Equal("archlinux-test", h.Distro.Id)
	assert.Equal(evergreen.ProviderNameEc2Spot, h.Provider)
	assert.Equal(evergreen.ProviderNameEc2Spot, h.Distro.Provider)
	assert.Equal(distro.BootstrapMethodNone, h.Distro.BootstrapSettings.Method, "host provisioning should be set to none by default")

	ec2Settings = &cloud.EC2ProviderSettings{}
	ec2Settings2 = &cloud.EC2ProviderSettings{}
	require.Len(h.Distro.ProviderSettingsList, 1)
	assert.NoError(ec2Settings.FromDistroSettings(h.Distro, ""))
	assert.Equal("ami-123456", ec2Settings.AMI)
	assert.Equal("mock_key", ec2Settings.KeyName)
	assert.Equal(true, ec2Settings.IsVpc)

	h, err = host.FindOneByIdOrTag(h.Id)
	assert.NoError(err)
	require.NotNil(h)
	ec2Settings2 = &cloud.EC2ProviderSettings{}
	require.Len(h.Distro.ProviderSettingsList, 1)
	assert.NoError(ec2Settings2.FromDistroSettings(h.Distro, ""))
	assert.Equal("ami-123456", ec2Settings2.AMI)
	assert.Equal("mock_key", ec2Settings2.KeyName)
	assert.Equal(true, ec2Settings2.IsVpc)

	// override some evergreen distro settings
	c = apimodels.CreateHost{
		Distro:              "archlinux-test",
		CloudProvider:       "ec2",
		NumHosts:            "1",
		Scope:               "task",
		SetupTimeoutSecs:    600,
		TeardownTimeoutSecs: 21600,
		Spot:                true,
		AWSKeyID:            "my_aws_key",
		AWSSecret:           "my_secret_key",
		Subnet:              "subnet-123456",
	}
	handler.createHost = c
	h, err = handler.sc.MakeIntentHost(handler.taskID, "", "", handler.createHost)
	assert.NoError(err)
	assert.NotNil(h)
	assert.Equal("archlinux-test", h.Distro.Id)
	assert.Equal(evergreen.ProviderNameEc2Spot, h.Provider)
	assert.Equal(evergreen.ProviderNameEc2Spot, h.Distro.Provider)
	assert.Equal(distro.BootstrapMethodNone, h.Distro.BootstrapSettings.Method, "host provisioning should be set to none by default")

	ec2Settings = &cloud.EC2ProviderSettings{}
	require.Len(h.Distro.ProviderSettingsList, 1)
	assert.NoError(ec2Settings.FromDistroSettings(h.Distro, ""))
	assert.Equal("ami-123456", ec2Settings.AMI)
	assert.Equal("my_aws_key", ec2Settings.AWSKeyID)
	assert.Equal("my_secret_key", ec2Settings.AWSSecret)
	assert.Equal("subnet-123456", ec2Settings.SubnetId)
	assert.Equal(true, ec2Settings.IsVpc)

	ec2Settings2 = &cloud.EC2ProviderSettings{}
	require.Len(h.Distro.ProviderSettingsList, 1)
	assert.NoError(ec2Settings2.FromDistroSettings(h.Distro, ""))
	assert.Equal("ami-123456", ec2Settings2.AMI)
	assert.Equal("my_aws_key", ec2Settings2.AWSKeyID)
	assert.Equal("my_secret_key", ec2Settings2.AWSSecret)
	assert.Equal("subnet-123456", ec2Settings2.SubnetId)
	assert.Equal(true, ec2Settings2.IsVpc)

	// bring your own ami
	c = apimodels.CreateHost{
		AMI:                 "ami-654321",
		CloudProvider:       "ec2",
		NumHosts:            "1",
		Scope:               "task",
		SetupTimeoutSecs:    600,
		TeardownTimeoutSecs: 21600,
		Spot:                true,
		AWSKeyID:            "my_aws_key",
		AWSSecret:           "my_secret_key",
		InstanceType:        "t1.micro",
		Subnet:              "subnet-123456",
		SecurityGroups:      []string{"1234"},
	}
	handler.createHost = c
	h, err = handler.sc.MakeIntentHost(handler.taskID, "", "", handler.createHost)
	require.NoError(err)
	require.NotNil(h)
	assert.Equal("", h.Distro.Id)
	assert.Equal(evergreen.ProviderNameEc2Spot, h.Provider)
	assert.Equal(evergreen.ProviderNameEc2Spot, h.Distro.Provider)
	assert.Equal(distro.BootstrapMethodNone, h.Distro.BootstrapSettings.Method, "host provisioning should be set to none by default")

	ec2Settings2 = &cloud.EC2ProviderSettings{}
	require.Len(h.Distro.ProviderSettingsList, 1)
	assert.NoError(ec2Settings2.FromDistroSettings(h.Distro, ""))
	assert.Equal("ami-654321", ec2Settings2.AMI)
	assert.Equal("my_aws_key", ec2Settings2.AWSKeyID)
	assert.Equal("my_secret_key", ec2Settings2.AWSSecret)
	assert.Equal("subnet-123456", ec2Settings2.SubnetId)
	assert.Equal(true, ec2Settings2.IsVpc)

	// with multiple regions
	require.Len(d.ProviderSettingsList, 1)
	doc2 := d.ProviderSettingsList[0].Copy().Set(birch.EC.String("region", "us-west-1")).Set(birch.EC.String("ami", "ami-987654"))
	d.ProviderSettingsList = append(d.ProviderSettingsList, doc2)
	require.NoError(d.Update())
	c = apimodels.CreateHost{
		Distro:              "archlinux-test",
		CloudProvider:       "ec2",
		NumHosts:            "1",
		Scope:               "task",
		SetupTimeoutSecs:    600,
		TeardownTimeoutSecs: 21600,
		KeyName:             "mock_key",
	}
	handler.createHost = c
	h, err = handler.sc.MakeIntentHost(handler.taskID, "", "", handler.createHost)
	assert.NoError(err)
	assert.NotNil(h)
	assert.Equal("archlinux-test", h.Distro.Id)
	require.Len(h.Distro.ProviderSettingsList, 1)
	ec2Settings2 = &cloud.EC2ProviderSettings{}
	assert.NoError(ec2Settings2.FromDistroSettings(h.Distro, "us-east-1"))
	assert.Equal(ec2Settings2.AMI, "ami-123456")

	handler.createHost.Region = "us-west-1"
	h, err = handler.sc.MakeIntentHost(handler.taskID, "", "", handler.createHost)
	assert.NoError(err)
	assert.NotNil(h)
	assert.Equal("archlinux-test", h.Distro.Id)
	require.Len(h.Distro.ProviderSettingsList, 1)
	ec2Settings2 = &cloud.EC2ProviderSettings{}
	assert.NoError(ec2Settings2.FromDistroSettings(h.Distro, "us-west-1"))
	assert.Equal(ec2Settings2.AMI, "ami-987654")
}

func TestHostCreateDocker(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	require.NoError(db.ClearCollections(distro.Collection, host.Collection, task.Collection, evergreen.ConfigCollection))
	handler := hostCreateHandler{
		sc: &data.DBConnector{},
	}
	pool := evergreen.ContainerPool{Distro: "parent-distro", Id: "test-pool", MaxContainers: 2}
	poolConfig := evergreen.ContainerPoolsConfig{Pools: []evergreen.ContainerPool{pool}}
	settings := evergreen.Settings{ContainerPools: poolConfig}
	assert.NoError(evergreen.UpdateConfig(&settings))
	parent := distro.Distro{
		Id:       "parent-distro",
		Provider: evergreen.ProviderNameDockerMock,
		HostAllocatorSettings: distro.HostAllocatorSettings{
			MaximumHosts: 3,
		},
		ContainerPool: pool.Id,
	}
	require.NoError(parent.Insert())

	parentHost := &host.Host{
		Id:                    "host1",
		Host:                  "host",
		User:                  "user",
		Distro:                distro.Distro{Id: "parent-distro"},
		Status:                evergreen.HostRunning,
		HasContainers:         true,
		ContainerPoolSettings: &pool,
	}
	require.NoError(parentHost.Insert())

	d := distro.Distro{Id: "distro", Provider: evergreen.ProviderNameDockerMock, ContainerPool: "test-pool"}
	require.NoError(d.Insert())

	sampleTask := &task.Task{
		Id: handler.taskID,
	}
	require.NoError(sampleTask.Insert())
	c := apimodels.CreateHost{
		CloudProvider: apimodels.ProviderDocker,
		NumHosts:      "1",
		Distro:        "distro",
		Image:         "my-image",
		Command:       "echo hello",
	}
	c.Registry.Name = "myregistry"
	handler.createHost = c
	h, err := handler.sc.MakeIntentHost(handler.taskID, "", "", handler.createHost)
	assert.NoError(err)
	require.NotNil(h)
	assert.Equal("distro", h.Distro.Id)
	assert.Equal("my-image", h.DockerOptions.Image)
	assert.Equal("echo hello", h.DockerOptions.Command)
	assert.Equal("myregistry", h.DockerOptions.RegistryName)

	assert.Equal(200, handler.Run(context.Background()).Status())

	hosts, err := host.Find(db.Q{})
	assert.NoError(err)
	require.Len(hosts, 3)
	assert.Equal(h.DockerOptions.Command, hosts[1].DockerOptions.Command)
}

func TestGetDockerLogs(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	require.NoError(db.ClearCollections(distro.Collection, host.Collection, task.Collection, evergreen.ConfigCollection))
	handler := containerLogsHandler{
		sc: &data.MockConnector{},
	}
	pool := evergreen.ContainerPool{Distro: "parent-distro", Id: "test-pool", MaxContainers: 2}
	poolConfig := evergreen.ContainerPoolsConfig{Pools: []evergreen.ContainerPool{pool}}
	settings := evergreen.Settings{ContainerPools: poolConfig}
	assert.NoError(evergreen.UpdateConfig(&settings))

	parent := distro.Distro{
		Id:       "parent-distro",
		Provider: evergreen.ProviderNameMock,
		HostAllocatorSettings: distro.HostAllocatorSettings{
			MaximumHosts: 3,
		},
		ContainerPool: pool.Id,
	}
	require.NoError(parent.Insert())

	parentHost := &host.Host{
		Id:                    "host1",
		Host:                  "host",
		User:                  "user",
		Distro:                distro.Distro{Id: "parent-distro"},
		Status:                evergreen.HostRunning,
		HasContainers:         true,
		ContainerPoolSettings: &pool,
	}
	require.NoError(parentHost.Insert())

	d := distro.Distro{Id: "distro", Provider: evergreen.ProviderNameDockerMock, ContainerPool: "test-pool"}
	require.NoError(d.Insert())

	myTask := task.Task{
		Id:      "task-id",
		BuildId: "build-id",
	}
	require.NoError(myTask.Insert())
	c := apimodels.CreateHost{
		CloudProvider: apimodels.ProviderDocker,
		NumHosts:      "1",
		Distro:        "distro",
		Image:         "my-image",
		Command:       "echo hello",
	}
	h, err := handler.sc.MakeIntentHost("task-id", "", "", c)
	require.NoError(err)
	require.NotNil(h)
	assert.NotEmpty(h.ParentID)

	// invalid tail
	url := fmt.Sprintf("/hosts/%s/logs/output?tail=%s", h.Id, "invalid")
	request, err := http.NewRequest("GET", url, bytes.NewReader(nil))
	assert.NoError(err)
	options := map[string]string{"host_id": h.Id}

	request = gimlet.SetURLVars(request, options)
	assert.Error(handler.Parse(context.Background(), request))

	url = fmt.Sprintf("/hosts/%s/logs/output?tail=%s", h.Id, "-1")
	request, err = http.NewRequest("GET", url, bytes.NewReader(nil))
	assert.NoError(err)
	options = map[string]string{"host_id": h.Id}

	request = gimlet.SetURLVars(request, options)
	assert.Error(handler.Parse(context.Background(), request))

	// invalid Parse start time
	startTime := time.Now().Add(-time.Minute).String()
	url = fmt.Sprintf("/hosts/%s/logs/output?start_time=%s", h.Id, startTime)
	request, err = http.NewRequest("GET", url, bytes.NewReader(nil))
	assert.NoError(err)
	options = map[string]string{"host_id": h.Id}

	request = gimlet.SetURLVars(request, options)
	assert.Error(handler.Parse(context.Background(), request))

	// valid Parse
	startTime = time.Now().Add(-time.Minute).Format(time.RFC3339)
	endTime := time.Now().Format(time.RFC3339)
	url = fmt.Sprintf("/hosts/%s/logs/output?start_time=%s&end_time=%s&tail=10", h.Id, startTime, endTime)

	request, err = http.NewRequest("GET", url, bytes.NewReader(nil))
	assert.NoError(err)
	request = gimlet.SetURLVars(request, options)

	assert.NoError(handler.Parse(context.Background(), request))
	assert.Equal(h.Id, handler.host.Id)
	assert.Equal(startTime, handler.startTime)
	assert.Equal(endTime, handler.endTime)
	assert.Equal("10", handler.tail)

	// valid Run
	res := handler.Run(context.Background())
	require.NotNil(res)
	logs, ok := res.Data().(*bytes.Buffer)
	require.True(ok)
	assert.NoError(err)
	assert.Contains(logs.String(), "this is a log message")

}

func TestGetDockerStatus(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	require.NoError(db.ClearCollections(distro.Collection, host.Collection, task.Collection, evergreen.ConfigCollection))
	handler := containerStatusHandler{
		sc: &data.MockConnector{},
	}
	pool := evergreen.ContainerPool{Distro: "parent-distro", Id: "test-pool", MaxContainers: 2}
	poolConfig := evergreen.ContainerPoolsConfig{Pools: []evergreen.ContainerPool{pool}}
	settings := evergreen.Settings{ContainerPools: poolConfig}
	assert.NoError(evergreen.UpdateConfig(&settings))
	parent := distro.Distro{
		Id:       "parent-distro",
		Provider: evergreen.ProviderNameDockerMock,
		HostAllocatorSettings: distro.HostAllocatorSettings{
			MaximumHosts: 3,
		},
		ContainerPool: pool.Id,
	}
	require.NoError(parent.Insert())

	parentHost := &host.Host{
		Id:                    "host1",
		Host:                  "host",
		User:                  "user",
		Distro:                distro.Distro{Id: "parent-distro"},
		Status:                evergreen.HostRunning,
		HasContainers:         true,
		ContainerPoolSettings: &pool,
	}
	require.NoError(parentHost.Insert())

	d := distro.Distro{Id: "distro", Provider: evergreen.ProviderNameDockerMock, ContainerPool: "test-pool"}
	require.NoError(d.Insert())

	myTask := task.Task{
		Id:      "task-id",
		BuildId: "build-id",
	}
	require.NoError(myTask.Insert())
	c := apimodels.CreateHost{
		CloudProvider: apimodels.ProviderDocker,
		NumHosts:      "1",
		Distro:        "distro",
		Image:         "my-image",
		Command:       "echo hello",
	}
	h, err := handler.sc.MakeIntentHost("task-id", "", "", c)
	require.NoError(err)
	assert.NotEmpty(h.ParentID)

	url := fmt.Sprintf("/hosts/%s/logs/status", h.Id)
	options := map[string]string{"host_id": h.Id}

	request, err := http.NewRequest("GET", url, bytes.NewReader(nil))
	assert.NoError(err)
	request = gimlet.SetURLVars(request, options)

	assert.NoError(handler.Parse(context.Background(), request))
	require.NotNil(handler.host)
	assert.Equal(h.Id, handler.host.Id)

	// valid Run
	res := handler.Run(context.Background())
	require.NotNil(res)
	assert.Equal(http.StatusOK, res.Status())

	status, ok := res.Data().(*cloud.ContainerStatus)
	require.True(ok)
	require.NotNil(status)
	require.True(status.HasStarted)

}
