package route

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/evergreen-ci/birch"
	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/db"
	"github.com/evergreen-ci/evergreen/model/distro"
	"github.com/evergreen-ci/evergreen/model/event"
	"github.com/evergreen-ci/evergreen/model/host"
	"github.com/evergreen-ci/evergreen/model/user"
	"github.com/evergreen-ci/evergreen/rest/data"
	"github.com/evergreen-ci/evergreen/rest/model"
	"github.com/evergreen-ci/evergreen/testutil"
	"github.com/evergreen-ci/gimlet"
	"github.com/evergreen-ci/utility"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostPostHandler(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	require.NoError(db.ClearCollections(distro.Collection, host.Collection))

	config, err := evergreen.GetConfig()
	assert.NoError(err)
	config.Spawnhost.SpawnHostsPerUser = evergreen.DefaultMaxSpawnHostsPerUser
	doc := birch.NewDocument(
		birch.EC.String("ami", "ami-123"),
		birch.EC.String("region", evergreen.DefaultEC2Region),
	)
	d := &distro.Distro{
		Id:                   "distro",
		SpawnAllowed:         true,
		Provider:             evergreen.ProviderNameEc2OnDemand,
		ProviderSettingsList: []*birch.Document{doc},
	}
	require.NoError(d.Insert())
	assert.NoError(err)
	h := &hostPostHandler{
		settings: config,
		options: &model.HostRequestOptions{
			TaskID:   "task",
			DistroID: "distro",
			KeyName:  "keyname",
		},
	}
	h.sc = &data.MockConnector{}
	ctx := gimlet.AttachUser(context.Background(), &user.DBUser{Id: "user"})

	resp := h.Run(ctx)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.Status())
	h.options.UserData = "#!/bin/bash\necho my script"
	resp = h.Run(ctx)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.Status())
	h.options.InstanceTags = []host.Tag{
		host.Tag{
			Key:           "key",
			Value:         "value",
			CanBeModified: true,
		},
	}
	resp = h.Run(ctx)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.Status())

	d.Provider = evergreen.ProviderNameMock
	assert.NoError(d.Update())
	h.settings.Providers.AWS.AllowedInstanceTypes = append(h.settings.Providers.AWS.AllowedInstanceTypes, "test_instance_type")
	h.options.InstanceType = "test_instance_type"
	h.options.UserData = ""
	resp = h.Run(ctx)
	require.NotNil(resp)
	assert.Equal(http.StatusOK, resp.Status())

	assert.Len(h.sc.(*data.MockConnector).MockHostConnector.CachedHosts, 4)
	h0 := h.sc.(*data.MockConnector).MockHostConnector.CachedHosts[0]
	d0 := h0.Distro
	userdata, ok := d0.ProviderSettingsList[0].Lookup("user_data").StringValueOK()
	assert.False(ok)
	assert.Empty(userdata)
	assert.Empty(h0.InstanceTags)
	assert.Empty(h0.InstanceType)

	h1 := h.sc.(*data.MockConnector).MockHostConnector.CachedHosts[1]
	d1 := h1.Distro
	userdata, ok = d1.ProviderSettingsList[0].Lookup("user_data").StringValueOK()
	assert.True(ok)
	assert.Equal("#!/bin/bash\necho my script", userdata)
	assert.Empty(h1.InstanceTags)
	assert.Empty(h1.InstanceType)

	h2 := h.sc.(*data.MockConnector).MockHostConnector.CachedHosts[2]
	assert.Equal([]host.Tag{host.Tag{Key: "key", Value: "value", CanBeModified: true}}, h2.InstanceTags)
	assert.Empty(h2.InstanceType)

	h3 := h.sc.(*data.MockConnector).MockHostConnector.CachedHosts[3]
	assert.Equal("test_instance_type", h3.InstanceType)
}

func TestHostStopHandler(t *testing.T) {
	testutil.DisablePermissionsForTests()
	defer testutil.EnablePermissionsForTests()
	h := &hostStopHandler{
		sc:  &data.MockConnector{},
		env: evergreen.GetEnvironment(),
	}
	ctx := gimlet.AttachUser(context.Background(), &user.DBUser{Id: "user"})

	h.sc.(*data.MockConnector).MockHostConnector.CachedHosts = []host.Host{
		host.Host{
			Id:     "host-stopped",
			Status: evergreen.HostStopped,
		},
		host.Host{
			Id:     "host-provisioning",
			Status: evergreen.HostProvisioning,
		},
		host.Host{
			Id:     "host-running",
			Status: evergreen.HostRunning,
		},
	}

	h.hostID = "host-stopped"
	resp := h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusBadRequest, resp.Status())

	h.hostID = "host-provisioning"
	resp = h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusBadRequest, resp.Status())

	h.hostID = "host-running"
	h.subscriptionType = event.SlackSubscriberType
	resp = h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.Status())

	subscriptionConnector := h.sc.(*data.MockConnector).MockSubscriptionConnector
	assert.Len(t, subscriptionConnector.MockSubscriptions, 1)
}

func TestHostStartHandler(t *testing.T) {
	testutil.DisablePermissionsForTests()
	defer testutil.EnablePermissionsForTests()
	h := &hostStartHandler{
		sc:  &data.MockConnector{},
		env: evergreen.GetEnvironment(),
	}
	ctx := gimlet.AttachUser(context.Background(), &user.DBUser{Id: "user"})

	h.sc.(*data.MockConnector).MockHostConnector.CachedHosts = []host.Host{
		host.Host{
			Id:     "host-running",
			Status: evergreen.HostRunning,
		},
		host.Host{
			Id:     "host-stopped",
			Status: evergreen.HostStopped,
		},
	}

	h.hostID = "host-running"
	resp := h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusBadRequest, resp.Status())

	h.hostID = "host-stopped"
	h.subscriptionType = event.SlackSubscriberType
	resp = h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.Status())

	subscriptionConnector := h.sc.(*data.MockConnector).MockSubscriptionConnector
	assert.Len(t, subscriptionConnector.MockSubscriptions, 1)
}

func TestCreateVolumeHandler(t *testing.T) {
	assert.NoError(t, db.ClearCollections(host.VolumesCollection))
	h := &createVolumeHandler{
		sc:       &data.MockConnector{},
		env:      evergreen.GetEnvironment(),
		provider: evergreen.ProviderNameMock,
	}
	ctx := gimlet.AttachUser(context.Background(), &user.DBUser{Id: "user"})
	v := host.Volume{ID: "volume1", Size: 15, CreatedBy: "user"}
	assert.NoError(t, v.Insert())
	v = host.Volume{ID: "volume2", Size: 35, CreatedBy: "user"}
	assert.NoError(t, v.Insert())
	v = host.Volume{ID: "not-relevant", Size: 400, CreatedBy: "someone-else"}
	assert.NoError(t, v.Insert())

	h.env.Settings().Providers.AWS.MaxVolumeSizePerUser = 100
	h.env.Settings().Providers.AWS.Subnets = []evergreen.Subnet{
		{AZ: "us-east-1a", SubnetID: "123"},
	}
	v = host.Volume{ID: "volume-new", Size: 80, CreatedBy: "user"}
	h.volume = &v

	resp := h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusBadRequest, resp.Status())

	v.Size = 50
	resp = h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.Status())
}

func TestDeleteVolumeHandler(t *testing.T) {
	assert.NoError(t, db.ClearCollections(host.VolumesCollection))
	h := &deleteVolumeHandler{
		sc:       &data.MockConnector{},
		env:      evergreen.GetEnvironment(),
		provider: evergreen.ProviderNameMock,
	}
	ctx := gimlet.AttachUser(context.Background(), &user.DBUser{Id: "user"})

	h.sc.(*data.MockConnector).MockHostConnector = data.MockHostConnector{
		CachedHosts: []host.Host{
			host.Host{
				Id:        "my-host",
				StartedBy: "user",
				Status:    evergreen.HostRunning,
				Volumes: []host.VolumeAttachment{
					{
						VolumeID:   "my-volume",
						DeviceName: "my-device",
					},
				},
			},
		},
		CachedVolumes: []host.Volume{
			host.Volume{
				ID:        "my-volume",
				CreatedBy: "user",
			},
		},
	}

	h.VolumeID = "my-volume"
	resp := h.Run(ctx)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusBadRequest, resp.Status())
}

func TestAttachVolumeHandler(t *testing.T) {
	assert.NoError(t, db.ClearCollections(host.VolumesCollection))
	h := &attachVolumeHandler{
		sc:  &data.MockConnector{},
		env: evergreen.GetEnvironment(),
	}
	ctx := gimlet.AttachUser(context.Background(), &user.DBUser{Id: "user"})
	h.sc.(*data.MockConnector).MockHostConnector.CachedHosts = []host.Host{
		host.Host{
			Id:        "my-host",
			Status:    evergreen.HostRunning,
			StartedBy: "user",
			Zone:      "us-east-1c",
		},
		host.Host{
			Id: "different-host",
		},
	}

	// no volume
	v := &host.VolumeAttachment{DeviceName: "my-device"}
	jsonBody, err := json.Marshal(v)
	assert.NoError(t, err)
	buffer := bytes.NewBuffer(jsonBody)

	r, err := http.NewRequest("GET", "/hosts/my-host/attach", buffer)
	assert.NoError(t, err)
	r = gimlet.SetURLVars(r, map[string]string{"host_id": "my-host"})

	assert.Error(t, h.Parse(ctx, r))

	// wrong availability zone
	v.VolumeID = "my-volume"
	h.sc.(*data.MockConnector).MockHostConnector.CachedVolumes = []host.Volume{
		host.Volume{
			ID: v.VolumeID,
		},
	}

	jsonBody, err = json.Marshal(v)
	assert.NoError(t, err)
	buffer = bytes.NewBuffer(jsonBody)

	r, err = http.NewRequest("GET", "/hosts/my-host/attach", buffer)
	assert.NoError(t, err)
	r = gimlet.SetURLVars(r, map[string]string{"host_id": "my-host"})

	assert.NoError(t, h.Parse(ctx, r))

	require.NotNil(t, h.attachment)
	assert.Equal(t, h.attachment.VolumeID, "my-volume")
	assert.Equal(t, h.attachment.DeviceName, "my-device")

	resp := h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusBadRequest, resp.Status())
}

func TestDetachVolumeHandler(t *testing.T) {
	h := &detachVolumeHandler{
		sc:  &data.MockConnector{},
		env: evergreen.GetEnvironment(),
	}
	ctx := gimlet.AttachUser(context.Background(), &user.DBUser{Id: "user"})

	h.sc.(*data.MockConnector).MockHostConnector.CachedHosts = []host.Host{
		host.Host{
			Id:        "my-host",
			StartedBy: "user",
			Status:    evergreen.HostRunning,
			Volumes: []host.VolumeAttachment{
				{
					VolumeID:   "my-volume",
					DeviceName: "my-device",
				},
			},
		},
	}

	v := host.VolumeAttachment{VolumeID: "not-a-volume"}
	jsonBody, err := json.Marshal(v)
	assert.NoError(t, err)
	buffer := bytes.NewBuffer(jsonBody)

	r, err := http.NewRequest("GET", "/hosts/my-host/detach", buffer)
	assert.NoError(t, err)
	r = gimlet.SetURLVars(r, map[string]string{"host_id": "my-host"})

	assert.NoError(t, h.Parse(ctx, r))
	resp := h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusNotFound, resp.Status())
}

func TestModifyVolumeHandler(t *testing.T) {
	h := &modifyVolumeHandler{
		sc:   &data.MockConnector{},
		env:  evergreen.GetEnvironment(),
		opts: &model.VolumeModifyOptions{},
	}
	h.env.Settings().Providers.AWS.MaxVolumeSizePerUser = 200
	h.env.Settings().Spawnhost.UnexpirableVolumesPerUser = 1
	h.sc.(*data.MockConnector).MockHostConnector = data.MockHostConnector{
		CachedVolumes: []host.Volume{
			{
				ID:               "volume1",
				CreatedBy:        "user",
				Size:             64,
				AvailabilityZone: evergreen.DefaultEBSAvailabilityZone,
			},
		},
	}

	// parse request
	opts := &model.VolumeModifyOptions{Size: 20, NewName: "my-favorite-volume"}
	jsonBody, err := json.Marshal(opts)
	assert.NoError(t, err)
	buffer := bytes.NewBuffer(jsonBody)
	r, err := http.NewRequest("", "", buffer)
	assert.NoError(t, err)
	r = gimlet.SetURLVars(r, map[string]string{"volume_id": "volume1"})
	assert.NoError(t, h.Parse(context.Background(), r))
	assert.Equal(t, "volume1", h.volumeID)
	assert.Equal(t, 20, h.opts.Size)
	assert.Equal(t, "my-favorite-volume", h.opts.NewName)

	h.provider = evergreen.ProviderNameMock
	h.opts = &model.VolumeModifyOptions{}
	// another user
	ctx := gimlet.AttachUser(context.Background(), &user.DBUser{Id: "different-user"})
	resp := h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusUnauthorized, resp.Status())

	// volume's owner
	ctx = gimlet.AttachUser(context.Background(), &user.DBUser{Id: "user"})
	resp = h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.Status())

	// resize
	h.opts.Size = 200
	resp = h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.Status())

	// resize, exceeding max size
	h.opts = &model.VolumeModifyOptions{Size: 500}
	resp = h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusBadRequest, resp.Status())

	// set expiration
	h.opts = &model.VolumeModifyOptions{Expiration: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)}
	resp = h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.Status())

	// no expiration
	h.opts = &model.VolumeModifyOptions{NoExpiration: true}
	resp = h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.Status())

	// has expiration
	h.opts = &model.VolumeModifyOptions{HasExpiration: true}
	resp = h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.Status())
}

func TestGetVolumesHandler(t *testing.T) {
	h := &getVolumesHandler{
		sc: &data.MockConnector{},
	}
	ctx := gimlet.AttachUser(context.Background(), &user.DBUser{Id: "user"})

	h1 := host.Host{
		Id:        "has-a-volume",
		StartedBy: "user",
		Volumes: []host.VolumeAttachment{
			{VolumeID: "volume1", DeviceName: "/dev/sdf4"},
		},
	}

	h.sc.(*data.MockConnector).MockHostConnector = data.MockHostConnector{
		CachedHosts: []host.Host{h1},
		CachedVolumes: []host.Volume{
			{
				ID:               "volume1",
				Host:             "has-a-volume",
				CreatedBy:        "user",
				Type:             evergreen.DefaultEBSType,
				Size:             64,
				AvailabilityZone: evergreen.DefaultEBSAvailabilityZone,
			},
			{
				ID:               "volume2",
				CreatedBy:        "user",
				Type:             evergreen.DefaultEBSType,
				Size:             36,
				AvailabilityZone: evergreen.DefaultEBSAvailabilityZone,
			},
			{
				ID:        "volume3",
				CreatedBy: "different-user",
			},
		},
	}

	resp := h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.Status())

	volumes, ok := resp.Data().([]model.APIVolume)
	assert.True(t, ok)
	require.Len(t, volumes, 2)

	for _, v := range volumes {
		assert.Equal(t, "user", utility.FromStringPtr(v.CreatedBy))
		assert.Equal(t, evergreen.DefaultEBSType, utility.FromStringPtr(v.Type))
		assert.Equal(t, evergreen.DefaultEBSAvailabilityZone, utility.FromStringPtr(v.AvailabilityZone))
		if utility.FromStringPtr(v.ID) == "volume1" {
			assert.Equal(t, h1.Id, utility.FromStringPtr(v.HostID))
			assert.Equal(t, h1.Volumes[0].DeviceName, utility.FromStringPtr(v.DeviceName))
			assert.Equal(t, v.Size, 64)
		} else {
			assert.Empty(t, utility.FromStringPtr(v.HostID))
			assert.Empty(t, utility.FromStringPtr(v.DeviceName))
			assert.Equal(t, v.Size, 36)
		}
	}
}

func TestGetVolumeByIDHandler(t *testing.T) {
	h := &getVolumeByIDHandler{
		sc: &data.MockConnector{},
	}
	ctx := gimlet.AttachUser(context.Background(), &user.DBUser{Id: "user"})

	h1 := host.Host{
		Id:        "has-a-volume",
		StartedBy: "user",
		Volumes: []host.VolumeAttachment{
			{VolumeID: "volume1", DeviceName: "/dev/sdf4"},
		},
	}

	h.sc.(*data.MockConnector).MockHostConnector = data.MockHostConnector{
		CachedHosts: []host.Host{h1},
		CachedVolumes: []host.Volume{
			{
				ID:               "volume1",
				Host:             "has-a-volume",
				CreatedBy:        "user",
				Type:             evergreen.DefaultEBSType,
				Size:             64,
				AvailabilityZone: evergreen.DefaultEBSAvailabilityZone,
			},
		},
	}
	r, err := http.NewRequest("GET", "/volumes/volume1", nil)
	assert.NoError(t, err)
	r = gimlet.SetURLVars(r, map[string]string{"volume_id": "volume1"})
	assert.NoError(t, h.Parse(ctx, r))

	resp := h.Run(ctx)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.Status())

	v, ok := resp.Data().(*model.APIVolume)
	assert.True(t, ok)
	require.NotNil(t, v)
	assert.Equal(t, "user", utility.FromStringPtr(v.CreatedBy))
	assert.Equal(t, evergreen.DefaultEBSType, utility.FromStringPtr(v.Type))
	assert.Equal(t, evergreen.DefaultEBSAvailabilityZone, utility.FromStringPtr(v.AvailabilityZone))
	assert.Equal(t, h1.Id, utility.FromStringPtr(v.HostID))
	assert.Equal(t, h1.Volumes[0].DeviceName, utility.FromStringPtr(v.DeviceName))
	assert.Equal(t, v.Size, 64)
}

func TestMakeSpawnHostSubscription(t *testing.T) {
	user := &user.DBUser{
		EmailAddress: "evergreen@mongodb.com",
		Settings: user.UserSettings{
			SlackUsername: "mci",
		},
	}
	_, err := makeSpawnHostSubscription("id", "non-existent", user)
	assert.Error(t, err)

	sub, err := makeSpawnHostSubscription("id", event.SlackSubscriberType, user)
	assert.NoError(t, err)
	assert.Equal(t, event.ResourceTypeHost, utility.FromStringPtr(sub.ResourceType))
	assert.Len(t, sub.Selectors, 1)
	assert.Equal(t, event.SlackSubscriberType, utility.FromStringPtr(sub.Subscriber.Type))
	assert.Equal(t, "@mci", sub.Subscriber.Target)
}
