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
	Name:       "db",
	Protocol:   core.ProtocolTCP,
	Port:       3306,
	TargetPort: intstr.FromString("db"),
}

func (c *Controller) ensureService(pxc *api.PerconaXtraDB) (kutil.VerbType, error) {
	// Check if service name exists
	if err := c.checkService(pxc, pxc.ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	// create database Service
	vt, err := c.createService(pxc)
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			pxc,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s Service",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) checkService(pxc *api.PerconaXtraDB, serviceName string) error {
	service, err := c.Client.CoreV1().Services(pxc.Namespace).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if service.Labels[api.LabelDatabaseKind] != api.ResourceKindPerconaXtraDB ||
		service.Labels[api.LabelDatabaseName] != pxc.Name {
		return fmt.Errorf(`intended service "%v/%v" already exists`, pxc.Namespace, serviceName)
	}

	return nil
}

func (c *Controller) createService(pxc *api.PerconaXtraDB) (kutil.VerbType, error) {
	meta := metav1.ObjectMeta{
		Name:      pxc.OffshootName(),
		Namespace: pxc.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, pxc)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	_, ok, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = pxc.XtraDBLabels()
		in.Annotations = pxc.Spec.ServiceTemplate.Annotations

		in.Spec.Selector = pxc.XtraDBSelectors()
		in.Spec.Ports = ofst.MergeServicePorts(
			core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{defaultDBPort}),
			pxc.Spec.ServiceTemplate.Spec.Ports,
		)

		if pxc.Spec.ServiceTemplate.Spec.ClusterIP != "" {
			in.Spec.ClusterIP = pxc.Spec.ServiceTemplate.Spec.ClusterIP
		}
		if pxc.Spec.ServiceTemplate.Spec.Type != "" {
			in.Spec.Type = pxc.Spec.ServiceTemplate.Spec.Type
		}
		in.Spec.ExternalIPs = pxc.Spec.ServiceTemplate.Spec.ExternalIPs
		in.Spec.LoadBalancerIP = pxc.Spec.ServiceTemplate.Spec.LoadBalancerIP
		in.Spec.LoadBalancerSourceRanges = pxc.Spec.ServiceTemplate.Spec.LoadBalancerSourceRanges
		in.Spec.ExternalTrafficPolicy = pxc.Spec.ServiceTemplate.Spec.ExternalTrafficPolicy
		if pxc.Spec.ServiceTemplate.Spec.HealthCheckNodePort > 0 {
			in.Spec.HealthCheckNodePort = pxc.Spec.ServiceTemplate.Spec.HealthCheckNodePort
		}
		return in
	})
	return ok, err
}

func (c *Controller) ensureStatsService(pxc *api.PerconaXtraDB) (kutil.VerbType, error) {
	// return if monitoring is not prometheus
	if pxc.GetMonitoringVendor() != mona.VendorPrometheus {
		log.Infoln("spec.monitor.agent is not coreos-operator or builtin.")
		return kutil.VerbUnchanged, nil
	}

	// Check if statsService name exists
	if err := c.checkService(pxc, pxc.StatsService().ServiceName()); err != nil {
		return kutil.VerbUnchanged, err
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, pxc)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	// reconcile stats Service
	meta := metav1.ObjectMeta{
		Name:      pxc.StatsService().ServiceName(),
		Namespace: pxc.Namespace,
	}
	_, vt, err := core_util.CreateOrPatchService(c.Client, meta, func(in *core.Service) *core.Service {
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)
		in.Labels = pxc.StatsServiceLabels()
		in.Spec.Selector = pxc.OffshootSelectors()
		in.Spec.Ports = core_util.MergeServicePorts(in.Spec.Ports, []core.ServicePort{
			{
				Name:       api.PrometheusExporterPortName,
				Protocol:   core.ProtocolTCP,
				Port:       pxc.Spec.Monitor.Prometheus.Port,
				TargetPort: intstr.FromString(api.PrometheusExporterPortName),
			},
		})
		return in
	})
	if err != nil {
		return kutil.VerbUnchanged, err
	} else if vt != kutil.VerbUnchanged {
		c.recorder.Eventf(
			pxc,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %s stats service",
			vt,
		)
	}
	return vt, nil
}

func (c *Controller) createPerconaXtraDBGoverningService(pxc *api.PerconaXtraDB) (string, error) {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, pxc)
	if rerr != nil {
		return "", rerr
	}

	service := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pxc.GoverningServiceName(),
			Namespace: pxc.Namespace,
			//Labels:    pxc.OffshootLabels(),
			Labels: pxc.XtraDBLabels(),
			// 'tolerate-unready-endpoints' annotation is deprecated.
			// ref: https://github.com/kubernetes/kubernetes/pull/63742
			Annotations: map[string]string{
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
			},
		},
		Spec: core.ServiceSpec{
			Type:                     core.ServiceTypeClusterIP,
			ClusterIP:                core.ClusterIPNone,
			PublishNotReadyAddresses: true,
			Ports: []core.ServicePort{
				{
					Name: "db",
					Port: api.MySQLNodePort,
				},
			},
			Selector: pxc.XtraDBSelectors(),
		},
	}
	core_util.EnsureOwnerReference(&service.ObjectMeta, ref)

	_, err := c.Client.CoreV1().Services(pxc.Namespace).Create(service)
	if err != nil && !kerr.IsAlreadyExists(err) {
		return "", err
	}
	return service.Name, nil
}

func (c *Controller) createProxysqlService(pxc *api.PerconaXtraDB) (string, error) {
	ref, rerr := reference.GetReference(clientsetscheme.Scheme, pxc)
	if rerr != nil {
		return "", rerr
	}

	service := &core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pxc.ProxysqlServiceName(),
			Namespace: pxc.Namespace,
			Labels:    pxc.ProxysqlLabels(),
			// 'tolerate-unready-endpoints' annotation is deprecated.
			// ref: https://github.com/kubernetes/kubernetes/pull/63742
			Annotations: map[string]string{
				"service.alpha.kubernetes.io/tolerate-unready-endpoints": "true",
			},
		},
		Spec: core.ServiceSpec{
			Type:                     core.ServiceTypeClusterIP,
			ClusterIP:                core.ClusterIPNone,
			PublishNotReadyAddresses: true,
			Ports: []core.ServicePort{
				{
					Name: "mysql",
					Port: api.ProxysqlMySQLNodePort,
				},
			},
			Selector: pxc.ProxysqlSelectors(),
		},
	}
	core_util.EnsureOwnerReference(&service.ObjectMeta, ref)

	_, err := c.Client.CoreV1().Services(pxc.Namespace).Create(service)
	if err != nil && !kerr.IsAlreadyExists(err) {
		return "", err
	}
	return service.Name, nil
}
