package controller

import (
	"fmt"

	"github.com/appscode/go/log"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	core_util "kmodules.xyz/client-go/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"
)

var defaultDBPort = core.ServicePort{
	Name:       "mysql",
	Protocol:   core.ProtocolTCP,
	Port:       api.ProxySQLMySQLNodePort,
	TargetPort: intstr.FromInt(api.ProxySQLMySQLNodePort),
}

func (c *Controller) ensureService(proxysql *api.ProxySQL) (kutil.VerbType, error) {
	// Check if service name exists
	if err := c.checkService(proxysql, proxysql.ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	// create database Service
	vt, err := c.createService(proxysql)
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			proxysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Service",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) checkService(proxysql *api.ProxySQL, serviceName string) error {
	service, err := c.Client.CoreV1().Services(proxysql.Namespace).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if service.Labels[api.LabelDatabaseKind] != api.ResourceKindProxySQL ||
		service.Labels[api.LabelProxySQLName] != proxysql.Name {
		return fmt.Errorf(`intended service "%v/%v" already exists`, proxysql.Namespace, serviceName)
	}

	return nil
}

func (c *Controller) createService(proxysql *api.ProxySQL) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      proxysql.OffshootName(),
		Namespace: proxysql.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, proxysql)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	_, ok, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = proxysql.OffshootLabels()
		in.Annotations = proxysql.Spec.ServiceTemplate.Annotations

		in.Spec.Selector = proxysql.OffshootSelectors()
		in.Spec.Ports = ofst.MergeServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{defaultDBPort}),
			proxysql.Spec.ServiceTemplate.Spec.Ports,
		)

		if proxysql.Spec.ServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = proxysql.Spec.ServiceTemplate.Spec.ClusterIP
		}
		if proxysql.Spec.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = proxysql.Spec.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = proxysql.Spec.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = proxysql.Spec.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = proxysql.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = proxysql.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy
		if proxysql.Spec.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = proxysql.Spec.ServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	})
	return ok, err
}

func (c *Controller) ensureStatsService(proxysql *api.ProxySQL) (kutil.VerbType, error) {
	// return if monitoring is not prometheus
	if proxysql.GetMonitoringVendor() != mona.VendorPrometheus {
		log.Infoln("spec.monitor.agent is not coreos-operator or builtin.")
		return kutil.VerbUnchanged, nil
	}

	// Check if statsService name exists
	if err := c.checkService(proxysql, proxysql.StatsService().ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, proxysql)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	// reconcile stats Service
	meta := metav1.ObjectMeta{
		Name:      proxysql.StatsService().ServiceName(),
		Namespace: proxysql.Namespace,
	}
	_, vt, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = proxysql.StatsServiceLabels()
		in.Spec.Selector = proxysql.OffshootSelectors()
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       api.PrometheusExporterPortName,
				Protocol:   core.ProtocolTCP,
				Port:       proxysql.Spec.Monitor.Prometheus.Port,
				TargetPort: intstr.FromString(api.PrometheusExporterPortName),
			},
		})
		return in
	})
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			proxysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s stats service",
			vt,
		)
	}
	return vt, nil
}
