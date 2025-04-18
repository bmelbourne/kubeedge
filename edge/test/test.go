package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"github.com/kubeedge/api/apis/componentconfig/edgecore/v1alpha2"
	"github.com/kubeedge/beehive/pkg/core"
	beehiveContext "github.com/kubeedge/beehive/pkg/core/context"
	"github.com/kubeedge/beehive/pkg/core/model"
	"github.com/kubeedge/kubeedge/edge/pkg/common/message"
	"github.com/kubeedge/kubeedge/edge/pkg/common/modules"
)

const (
	name            = "testManager"
	edgedEndPoint   = "http://127.0.0.1:10255"
	EdgedPodHandler = "/pods"
)

// TODO move this files into /edge/pkg/dbtest @kadisi
func Register(t *v1alpha2.DBTest) {
	core.Register(&testManager{enable: t.Enable})
}

type testManager struct {
	enable bool
}

var _ core.Module = (*testManager)(nil)

func (testManager) Name() string {
	return name
}

func (testManager) Group() string {
	// return core.MetaGroup
	return modules.MetaGroup
}

func (tm *testManager) Enable() bool {
	return tm.enable
}

// Function to get the pods from Edged
func GetPodListFromEdged(w http.ResponseWriter) error {
	var pods v1.PodList
	var bytes io.Reader
	client := &http.Client{}
	t := time.Now()
	req, err := http.NewRequest(http.MethodGet, edgedEndPoint+EdgedPodHandler, bytes)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		klog.Errorf("Frame HTTP request failed: %v", err)
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		klog.Errorf("Sending HTTP request failed: %v", err)
		return err
	}
	klog.Infof("%s %s %v in %v", req.Method, req.URL, resp.Status, time.Since(t))
	defer resp.Body.Close()
	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Errorf("HTTP Response reading has failed: %v", err)
		return err
	}
	err = json.Unmarshal(contents, &pods)
	if err != nil {
		klog.Errorf("Json Unmarshal has failed: %v", err)
		return err
	}
	respBody, err := json.Marshal(pods)
	if err != nil {
		klog.Errorf("Json Marshal failed: %v", err)
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(respBody); err != nil {
		return fmt.Errorf("failed to write response, err: %v", err)
	}

	return nil
}

// podHandler handles Get/Add/Delete deployment list.
func (*testManager) podHandler(w http.ResponseWriter, req *http.Request) {
	var operation string
	var p v1.Pod
	if req.Method == http.MethodGet {
		err := GetPodListFromEdged(w)
		if err != nil {
			klog.Errorf("Get podlist from Edged has failed: %v", err)
		}
	} else if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			klog.Errorf("read body error %v", err)
			if _, err := w.Write([]byte("read request body error")); err != nil {
				// TODO: handle error
				klog.Error(err)
			}
		}
		klog.Infof("request body is %s", string(body))
		if err = json.Unmarshal(body, &p); err != nil {
			klog.Errorf("unmarshal request body error %v", err)
			if _, err := w.Write([]byte("unmarshal request body error")); err != nil {
				// TODO: handle error
				klog.Error(err)
			}
		}

		switch req.Method {
		case http.MethodPost:
			operation = model.InsertOperation
		case http.MethodDelete:
			operation = model.DeleteOperation
		case http.MethodPut:
			operation = model.UpdateOperation
		}

		ns := v1.NamespaceDefault
		if p.Namespace != "" {
			ns = p.Namespace
		}
		msgReq := message.BuildMsg("resource", string(p.UID), "edgecontroller", ns+"/pod/"+p.Name, operation, p)
		beehiveContext.Send(modules.MetaManagerModuleName, *msgReq)
		klog.Infof("send message to metaManager is %+v\n", msgReq)
	}
}

// Function to handle device addition and removal from the edgenode
func (*testManager) deviceHandler(w http.ResponseWriter, req *http.Request) {
	var operation string
	var Content interface{}

	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			klog.Errorf("read body error %v", err)
			if _, err := w.Write([]byte("read request body error")); err != nil {
				// TODO: handle error
				klog.Error(err)
			}
		}
		klog.Infof("request body is %s\n", string(body))
		err = json.Unmarshal(body, &Content)
		if err != nil {
			klog.Errorf("unmarshal request body error %v", err)
			if _, err := w.Write([]byte("unmarshal request body error")); err != nil {
				// TODO: handle error
				klog.Error(err)
			}
		}
		switch req.Method {
		case http.MethodPost:
			operation = model.InsertOperation
		case http.MethodDelete:
			operation = model.DeleteOperation
		case http.MethodPut:
			operation = model.UpdateOperation
		}
		msgReq := message.BuildMsg("edgehub", "", "edgemgr", "membership", operation, Content)
		beehiveContext.Send("twin", *msgReq)
		klog.Infof("send message to twingrp is %+v\n", msgReq)
	}
}

func (*testManager) secretHandler(w http.ResponseWriter, req *http.Request) {
	var operation string
	var p v1.Secret
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			klog.Errorf("read body error %v", err)
			if _, err := w.Write([]byte("read request body error")); err != nil {
				// TODO: handle error
				klog.Error(err)
			}
		}
		klog.Infof("request body is %s\n", string(body))
		if err = json.Unmarshal(body, &p); err != nil {
			klog.Errorf("unmarshal request body error %v", err)
			if _, err := w.Write([]byte("unmarshal request body error")); err != nil {
				// TODO: handle error
				klog.Error(err)
			}
		}

		switch req.Method {
		case http.MethodPost:
			operation = model.InsertOperation
		case http.MethodDelete:
			operation = model.DeleteOperation
		case http.MethodPut:
			operation = model.UpdateOperation
		}

		msgReq := message.BuildMsg("edgehub", string(p.UID), "test", "fakeNamespace/secret/"+string(p.UID), operation, p)
		beehiveContext.Send(modules.MetaManagerModuleName, *msgReq)
		klog.Infof("send message to metaManager is %+v\n", msgReq)
	}
}

func (*testManager) configmapHandler(w http.ResponseWriter, req *http.Request) {
	var operation string
	var p v1.ConfigMap
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			klog.Errorf("read body error %v", err)
			if _, err := w.Write([]byte("read request body error")); err != nil {
				// TODO: handle error
				klog.Error(err)
			}
		}
		klog.Infof("request body is %s\n", string(body))
		if err = json.Unmarshal(body, &p); err != nil {
			klog.Errorf("unmarshal request body error %v", err)
			if _, err := w.Write([]byte("unmarshal request body error")); err != nil {
				// TODO: handle error
				klog.Error(err)
			}
		}

		switch req.Method {
		case http.MethodPost:
			operation = model.InsertOperation
		case http.MethodDelete:
			operation = model.DeleteOperation
		case http.MethodPut:
			operation = model.UpdateOperation
		}

		msgReq := message.BuildMsg("edgehub", string(p.UID), "test", "fakeNamespace/configmap/"+string(p.UID), operation, p)
		beehiveContext.Send(modules.MetaManagerModuleName, *msgReq)
		klog.Infof("send message to metaManager is %+v\n", msgReq)
	}
}

func (tm *testManager) Start() {
	http.HandleFunc("/pods", tm.podHandler)
	http.HandleFunc("/configmap", tm.configmapHandler)
	http.HandleFunc("/secret", tm.secretHandler)
	http.HandleFunc("/devices", tm.deviceHandler)
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		klog.Errorf("ListenAndServe: %v", err)
	}
}
