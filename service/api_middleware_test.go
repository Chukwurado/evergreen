package service

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/db"
	"github.com/evergreen-ci/evergreen/model/host"
	"github.com/evergreen-ci/evergreen/model/task"
	"github.com/evergreen-ci/gimlet"
	. "github.com/smartystreets/goconvey/convey"
)

func TestCheckHostWrapper(t *testing.T) {
	h1 := host.Host{
		Id:          "h1",
		Secret:      "swordfish",
		RunningTask: "t1",
	}
	t1 := task.Task{
		Id:     "t1",
		Secret: "password",
		HostId: "h1",
	}

	env := evergreen.GetEnvironment()
	queue := env.LocalQueue()

	Convey("With a simple checkTask and checkHost-wrapped route", t, func() {
		if err := db.ClearCollections(host.Collection, task.Collection); err != nil {
			t.Fatalf("clearing db: %v", err)
		}

		as, err := NewAPIServer(env, queue)
		if err != nil {
			t.Fatalf("creating test API server: %v", err)
		}
		var (
			retrievedTask *task.Task
			retrievedHost *host.Host
		)

		app := gimlet.NewApp()
		app.NoVersions = true
		app.AddRoute("/{taskId}/").Handler(as.requireTaskStrict(as.requireHost(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				retrievedTask = GetTask(r)
				retrievedHost = GetHost(r)
				gimlet.WriteJSON(w, nil)
			}),
		))).Get()

		root, err := app.Handler()
		if err != nil {
			t.Fatalf("creating test handler server: %v", err)
		}

		Convey("and documents representing a proper host-task relationship", func() {
			So(t1.Insert(), ShouldBeNil)
			So(h1.Insert(), ShouldBeNil)

			w := httptest.NewRecorder()
			r, err := http.NewRequest("GET", "/t1/", nil)
			if err != nil {
				t.Fatalf("building request: %v", err)
			}

			Convey("a request without proper task fields should fail", func() {
				root.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusConflict)

				Convey("and attach nothing to the context", func() {
					So(retrievedTask, ShouldBeNil)
					So(retrievedHost, ShouldBeNil)
				})
			})
			Convey("a request with proper task fields but no host fields should not pass", func() {
				r.Header.Add(evergreen.TaskSecretHeader, t1.Secret)
				root.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusBadRequest)

				Convey("and attach nothing to the context", func() {
					So(retrievedTask, ShouldBeNil)
					So(retrievedHost, ShouldBeNil)
				})
			})
			Convey("a request with proper task fields and host fields should pass", func() {
				r.Header.Add(evergreen.TaskSecretHeader, t1.Secret)
				r.Header.Add(evergreen.HostHeader, h1.Id)
				r.Header.Add(evergreen.HostSecretHeader, h1.Secret)
				root.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)

				Convey("and attach the task and host to the context", func() {
					So(retrievedTask, ShouldNotBeNil)
					So(retrievedHost, ShouldNotBeNil)
					So(retrievedHost.Id, ShouldEqual, h1.Id)
					Convey("with an updated LastCommunicationTime", func() {
						So(retrievedHost.LastCommunicationTime, ShouldHappenWithin, time.Second, time.Now())
					})
				})
			})
			Convey("a request on a host should not need a new agent or monitor", func() {
				So(h1.SetNeedsNewAgent(true), ShouldBeNil)
				So(h1.SetNeedsNewAgentMonitor(true), ShouldBeNil)

				r.Header.Add(evergreen.TaskSecretHeader, t1.Secret)
				r.Header.Add(evergreen.HostHeader, h1.Id)
				r.Header.Add(evergreen.HostSecretHeader, h1.Secret)
				root.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)

				dbHost, err := host.FindOneId(h1.Id)
				So(err, ShouldBeNil)
				So(dbHost.NeedsNewAgent, ShouldBeFalse)
				So(dbHost.NeedsNewAgentMonitor, ShouldBeFalse)
			})
			Convey("a request with the wrong host secret should fail", func() {
				r.Header.Add(evergreen.TaskSecretHeader, t1.Secret)
				r.Header.Add(evergreen.HostHeader, h1.Id)
				r.Header.Add(evergreen.HostSecretHeader, "bad thing!!!")
				root.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusUnauthorized)
				msg, _ := ioutil.ReadAll(w.Body)
				So(string(msg), ShouldContainSubstring, "secret")

				Convey("and attach nothing to the context", func() {
					So(retrievedTask, ShouldBeNil)
					So(retrievedHost, ShouldBeNil)
				})
			})
		})
		Convey("and documents representing a mismatched host-task relationship", func() {
			h2 := host.Host{
				Id:          "h2",
				Secret:      "swordfish",
				RunningTask: "t29",
			}
			t2 := task.Task{
				Id:     "t2",
				Secret: "password",
				HostId: "h50",
			}
			So(t2.Insert(), ShouldBeNil)
			So(h2.Insert(), ShouldBeNil)

			w := httptest.NewRecorder()
			r, err := http.NewRequest("GET", "/t2/", nil)
			if err != nil {
				t.Fatalf("building request: %v", err)
			}

			Convey("a request with proper task fields and host fields should fail", func() {
				r.Header.Add(evergreen.TaskSecretHeader, t2.Secret)
				r.Header.Add(evergreen.HostHeader, h2.Id)
				r.Header.Add(evergreen.HostSecretHeader, h2.Secret)
				root.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusConflict)
				msg, _ := ioutil.ReadAll(w.Body)
				So(string(msg), ShouldContainSubstring, "should be running")

				Convey("and attach the and host to the context", func() {
					So(retrievedTask, ShouldBeNil)
					So(retrievedHost, ShouldBeNil)
				})
			})
		})
	})
	Convey("With a requireTask and requireHost-wrapped route using URL params", t, func() {
		if err := db.ClearCollections(host.Collection, task.Collection); err != nil {
			t.Fatalf("clearing db: %v", err)
		}

		as, err := NewAPIServer(env, queue)
		if err != nil {
			t.Fatalf("creating test API server: %v", err)
		}

		var (
			retrievedTask *task.Task
			retrievedHost *host.Host
		)

		app := gimlet.NewApp()
		app.NoVersions = true
		app.AddRoute("/{taskId}/{hostId}").Handler(as.requireTaskStrict(as.requireHost(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				retrievedTask = GetTask(r)
				retrievedHost = GetHost(r)

				gimlet.WriteJSON(w, nil)
			}),
		))).Get()

		root, err := app.Handler()
		if err != nil {
			t.Fatalf("creating test API server: %v", err)
		}

		Convey("and documents representing a proper host-task relationship", func() {

			So(t1.Insert(), ShouldBeNil)
			So(h1.Insert(), ShouldBeNil)

			w := httptest.NewRecorder()
			r, err := http.NewRequest("GET", "/t1/h1", nil)
			if err != nil {
				t.Fatalf("building request: %v", err)
			}

			Convey("a request with proper host params and fields should pass", func() {
				r.Header.Add(evergreen.TaskSecretHeader, t1.Secret)
				r.Header.Add(evergreen.HostSecretHeader, h1.Secret)
				root.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusOK)

				Convey("and attach the and host to the context", func() {
					So(retrievedTask, ShouldNotBeNil)
					So(retrievedHost, ShouldNotBeNil)
					So(retrievedHost.Id, ShouldEqual, h1.Id)
					Convey("with an updated LastCommunicationTime", func() {
						So(retrievedHost.LastCommunicationTime, ShouldHappenWithin, time.Second, time.Now())
					})
				})
			})
			Convey("a request with the wrong host secret should fail", func() {
				r.Header.Add(evergreen.TaskSecretHeader, t1.Secret)
				r.Header.Add(evergreen.HostSecretHeader, "bad thing!!!")
				root.ServeHTTP(w, r)
				So(w.Code, ShouldEqual, http.StatusUnauthorized)
				msg, _ := ioutil.ReadAll(w.Body)
				So(string(msg), ShouldContainSubstring, "secret")

				Convey("and attach nothing to the context", func() {
					So(retrievedTask, ShouldBeNil)
					So(retrievedHost, ShouldBeNil)
				})
			})
		})
	})
}
