package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	meta_util "kmodules.xyz/client-go/meta"
	"kmodules.xyz/monitoring-agent-api/agents"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
)

func (c *Controller) newMonitorController(pxc *api.PerconaXtraDB) (mona.Agent, error) {
	monitorSpec := pxc.Spec.Monitor

	if monitorSpec == nil {
		return nil, fmt.Errorf("MonitorSpec not found for PerconaXtraDB %v/%v in %v", pxc.Namespace, pxc.Name, pxc.Spec)
	}

	if monitorSpec.Prometheus != nil {
		return agents.New(monitorSpec.Agent, c.Client, c.ApiExtKubeClient, c.promClient), nil
	}

	return nil, fmt.Errorf("monitoring controller not found for PerconaXtraDB %v/%v in %v", pxc.Namespace, pxc.Name, monitorSpec)
}

func (c *Controller) addOrUpdateMonitor(pxc *api.PerconaXtraDB) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(pxc)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.CreateOrUpdate(pxc.StatsService(), pxc.Spec.Monitor)
}

func (c *Controller) deleteMonitor(pxc *api.PerconaXtraDB) (kutil.VerbType, error) {
	agent, err := c.newMonitorController(pxc)
	if err != nil {
		return kutil.VerbUnchanged, err
	}
	return agent.Delete(pxc.StatsService())
}

func (c *Controller) getOldAgent(pxc *api.PerconaXtraDB) mona.Agent {
	service, err := c.Client.CoreV1().Services(pxc.Namespace).Get(pxc.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return nil
	}
	oldAgentType, _ := meta_util.GetStringValue(service.Annotations, mona.KeyAgent)
	return agents.New(mona.AgentType(oldAgentType), c.Client, c.ApiExtKubeClient, c.promClient)
}

func (c *Controller) setNewAgent(pxc *api.PerconaXtraDB) error {
	service, err := c.Client.CoreV1().Services(pxc.Namespace).Get(pxc.StatsService().ServiceName(), metav1.GetOptions{})
	if err != nil {
		return err
	}
	_, _, err = core_util.PatchService(c.Client, service, func(in *core.Service) *core.Service {
		in.Annotations = core_util.UpsertMap(in.Annotations, map[string]string{
			mona.KeyAgent: string(pxc.Spec.Monitor.Agent),
		},
		)
		return in
	})
	return err
}

func (c *Controller) manageMonitor(pxc *api.PerconaXtraDB) error {
	oldAgent := c.getOldAgent(pxc)
	if pxc.Spec.Monitor != nil {
		if oldAgent != nil &&
			oldAgent.GetType() != pxc.Spec.Monitor.Agent {
			if _, err := oldAgent.Delete(pxc.StatsService()); err != nil {
				log.Errorf("error in deleting Prometheus agent. Reason: %s", err)
			}
		}
		if _, err := c.addOrUpdateMonitor(pxc); err != nil {
			return err
		}
		return c.setNewAgent(pxc)
	} else if oldAgent != nil {
		if _, err := oldAgent.Delete(pxc.StatsService()); err != nil {
			log.Errorf("error in deleting Prometheus agent. Reason: %s", err)
		}
	}
	return nil
}
