package controller

import (
	"fmt"
	"strings"

	"github.com/appscode/go/log"
	"github.com/appscode/go/types"
	"github.com/fatih/structs"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/reference"
	kutil "kmodules.xyz/client-go"
	app_util "kmodules.xyz/client-go/apps/v1"
	core_util "kmodules.xyz/client-go/core/v1"
	mona "kmodules.xyz/monitoring-agent-api/api/v1"
	ofst "kmodules.xyz/offshoot-api/api/v1"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/pkg/eventer"
)

type db interface {
	Replicas() int32
	PeerName(i int) string
	GetDatabaseSecretName() string
}

type workloadOptions struct {
	// App level options
	stsName   string
	labels    map[string]string
	selectors map[string]string

	// db container options
	conatainerName string
	image          string
	cmd            []string // cmd of `proxysql` container
	args           []string // args of `proxysql` container
	ports          []core.ContainerPort
	envList        []core.EnvVar // envList of `proxysql` container
	volumeMount    []core.VolumeMount
	configSource   *core.VolumeSource

	// monitor container
	monitorContainer *core.Container

	// pod Template level options
	replicas       *int32
	gvrSvcName     string
	podTemplate    *ofst.PodTemplateSpec
	pvcSpec        *core.PersistentVolumeClaimSpec
	initContainers []core.Container
	volume         []core.Volume // volumes to mount on stsPodTemplate
}

//func (c *Controller) ensureProxySQL(proxysql *api.ProxySQL) (kutil.VerbType, error) {
//	proxysqlVersion, err := c.ExtClient.CatalogV1alpha1().ProxySQLVersions().Get(string(proxysql.Spec.Version), metav1.GetOptions{})
//	if err != nil {
//		return kutil.VerbUnchanged, err
//	}
//
//	initContainers := append([]core.Container{
//		{
//			Name:            "remove-lost-found",
//			Image:           proxysqlVersion.Spec.InitContainer.Image,
//			ImagePullPolicy: core.PullIfNotPresent,
//			Command: []string{
//				"rm",
//				"-rf",
//				api.ProxySQLDataLostFoundPath,
//			},
//			VolumeMounts: []core.VolumeMount{
//				{
//					Name:      "data",
//					MountPath: api.ProxySQLDataMountPath,
//				},
//			},
//			Resources: proxysql.Spec.PodTemplate.Spec.Resources,
//		},
//	})
//
//	var cmds, args []string
//	var ports = []core.ContainerPort{
//		{
//			Name:          "mysql",
//			ContainerPort: api.MySQLNodePort,
//			Protocol:      core.ProtocolTCP,
//		},
//	}
//	if proxysql.Spec.PXC != nil {
//		cmds = []string{
//			"peer-finder",
//		}
//		userProvidedArgs := strings.Join(proxysql.Spec.PodTemplate.Spec.Args, " ")
//		args = []string{
//			fmt.Sprintf("-service=%s", c.GoverningService),
//			fmt.Sprintf("-on-start=/on-start.sh %s", userProvidedArgs),
//		}
//		ports = append(ports, []core.ContainerPort{
//			{
//				Name:          "sst",
//				ContainerPort: 4567,
//			},
//			{
//				Name:          "replication",
//				ContainerPort: 4568,
//			},
//		}...)
//	}
//
//	var volumes []core.Volume
//	var volumeMounts []core.VolumeMount
//
//	if proxysql.Spec.Init != nil && proxysql.Spec.Init.ScriptSource != nil {
//		volumes = append(volumes, core.Volume{
//			Name:         "initial-script",
//			VolumeSource: proxysql.Spec.Init.ScriptSource.VolumeSource,
//		})
//		volumeMounts = append(volumeMounts, core.VolumeMount{
//			Name:      "initial-script",
//			MountPath: api.ProxySQLInitDBMountPath,
//		})
//	}
//	proxysql.Spec.PodTemplate.Spec.ServiceAccountName = proxysql.OffshootName()
//
//	envList := []core.EnvVar{
//		{
//			Name: "MYSQL_USER",
//			ValueFrom: &core.EnvVarSource{
//				SecretKeyRef: &core.SecretKeySelector{
//					LocalObjectReference: core.LocalObjectReference{
//						Name: proxysql.Spec.DatabaseSecret.SecretName,
//					},
//					Key: api.ProxysqlUser,
//				},
//			},
//		},
//		{
//			Name: "MYSQL_PASSWORD",
//			ValueFrom: &core.EnvVarSource{
//				SecretKeyRef: &core.SecretKeySelector{
//					LocalObjectReference: core.LocalObjectReference{
//						Name: proxysql.Spec.DatabaseSecret.SecretName,
//					},
//					Key: api.ProxysqlPassword,
//				},
//			},
//		},
//	}
//	if proxysql.Spec.PXC != nil {
//		envList = append(envList, core.EnvVar{
//			Name:  "CLUSTER_NAME",
//			Value: proxysql.Name,
//		})
//	}
//
//	var monitorContainer core.Container
//	if proxysql.GetMonitoringVendor() == mona.VendorPrometheus {
//		monitorContainer = core.Container{
//			Name: "exporter",
//			Command: []string{
//				"/bin/sh",
//			},
//			Args: []string{
//				"-c",
//				// DATA_SOURCE_NAME=user:password@tcp(localhost:5555)/dbname
//				// ref: https://github.com/prometheus/mysqld_exporter#setting-the-mysql-servers-data-source-name
//				fmt.Sprintf(`export DATA_SOURCE_NAME="${MYSQL_ROOT_USERNAME:-}:${MYSQL_ROOT_PASSWORD:-}@(127.0.0.1:3306)/"
//						/bin/mysqld_exporter --web.listen-address=:%v --web.telemetry-path=%v %v`,
//					proxysql.Spec.Monitor.Prometheus.Port, proxysql.StatsService().Path(), strings.Join(proxysql.Spec.Monitor.Args, " ")),
//			},
//			Image: proxysqlVersion.Spec.Exporter.Image,
//			Ports: []core.ContainerPort{
//				{
//					Name:          api.PrometheusExporterPortName,
//					Protocol:      core.ProtocolTCP,
//					ContainerPort: proxysql.Spec.Monitor.Prometheus.Port,
//				},
//			},
//			Env:             proxysql.Spec.Monitor.Env,
//			Resources:       proxysql.Spec.Monitor.Resources,
//			SecurityContext: proxysql.Spec.Monitor.SecurityContext,
//		}
//	}
//
//	opts := workloadOptions{
//		stsName:          proxysql.OffshootName(),
//		labels:           proxysql.XtraDBLabels(),
//		selectors:        proxysql.XtraDBSelectors(),
//		conatainerName:   api.ResourceSingularProxySQL,
//		image:            proxysqlVersion.Spec.DB.Image,
//		args:             args,
//		cmd:              cmds,
//		ports:            ports,
//		envList:          envList,
//		initContainers:   initContainers,
//		gvrSvcName:       c.GoverningService,
//		podTemplate:      &proxysql.Spec.PodTemplate,
//		configSource:     proxysql.Spec.ConfigSource,
//		pvcSpec:          proxysql.Spec.Storage,
//		replicas:         proxysql.Spec.Replicas,
//		volume:           volumes,
//		volumeMount:      volumeMounts,
//		monitorContainer: &monitorContainer,
//	}
//
//	return c.ensureStatefulSet(proxysql, proxysql.Spec.UpdateStrategy, opts)
//}

func (c *Controller) ensureProxySQLNode(proxysql *api.ProxySQL) (kutil.VerbType, error) {
	proxysqlVersion, err := c.ExtClient.CatalogV1alpha1().ProxySQLVersions().Get(string(proxysql.Spec.Version), metav1.GetOptions{})
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	var ports = []core.ContainerPort{
		{
			Name:          "mysql",
			ContainerPort: api.ProxySQLMySQLNodePort,
			Protocol:      core.ProtocolTCP,
		},
		{
			Name:          api.ProxySQLAdminPortName,
			ContainerPort: api.ProxySQLAdminPort,
			Protocol:      core.ProtocolTCP,
		},
	}

	//var volumes []core.Volume
	//var volumeMounts []core.VolumeMount

	//volumeMounts = append(volumeMounts, core.VolumeMount{
	//	Name:      "data",
	//	MountPath: api.ProxysqlDataMountPath,
	//})
	//volumes = append(volumes, core.Volume{
	//	Name: "data",
	//	VolumeSource: core.VolumeSource{
	//		EmptyDir: &core.EmptyDirVolumeSource{},
	//	},
	//})

	proxysql.Spec.PodTemplate.Spec.ServiceAccountName = proxysql.OffshootName()

	var envList []core.EnvVar
	var peers []string
	var backendDB db

	switch backend := proxysql.Spec.Backend.Ref; backend.Kind {
	case api.ResourceKindPerconaXtraDB:
		backendDB, err = c.ExtClient.KubedbV1alpha1().PerconaXtraDBs(proxysql.Namespace).Get(backend.Name, metav1.GetOptions{})
	case api.ResourceKindMySQL:
		backendDB, err = c.ExtClient.KubedbV1alpha1().MySQLs(proxysql.Namespace).Get(backend.Name, metav1.GetOptions{})
		// TODO: add other cases for MySQL and MariaDB when they will be configured
	}
	if err != nil {
		return kutil.VerbUnchanged, err
	}

	for i := 0; i < int(backendDB.Replicas()); i += 1 {
		peers = append(peers, backendDB.PeerName(i))
	}

	envList = append(envList, []core.EnvVar{
		{
			Name: "MYSQL_ROOT_PASSWORD",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: backendDB.GetDatabaseSecretName(),
					},
					Key: MySQLPasswordKey,
				},
			},
		},
		{
			Name: "MYSQL_PROXY_USER",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: proxysql.Spec.ProxySQLSecret.SecretName,
					},
					Key: api.ProxySQLUserKey,
				},
			},
		},
		{
			Name: "MYSQL_PROXY_PASSWORD",
			ValueFrom: &core.EnvVarSource{
				SecretKeyRef: &core.SecretKeySelector{
					LocalObjectReference: core.LocalObjectReference{
						Name: proxysql.Spec.ProxySQLSecret.SecretName,
					},
					Key: api.ProxySQLPasswordKey,
				},
			},
		},
		{
			Name:  "PEERS",
			Value: strings.Join(peers, ","),
		},
		{
			Name:  "LOAD_BALANCE_MODE",
			Value: string(*proxysql.Spec.Mode),
		},
	}...)

	opts := workloadOptions{
		stsName:        proxysql.OffshootName(),
		labels:         proxysql.OffshootLabels(),
		selectors:      proxysql.OffshootSelectors(),
		conatainerName: api.ResourceSingularProxySQL,
		image:          proxysqlVersion.Spec.Proxysql.Image,
		args:           nil,
		cmd:            nil,
		ports:          ports,
		envList:        envList,
		initContainers: nil,
		gvrSvcName:     c.GoverningService,
		podTemplate:    &proxysql.Spec.PodTemplate,
		configSource:   proxysql.Spec.ConfigSource,
		pvcSpec:        proxysql.Spec.Storage,
		replicas:       proxysql.Spec.Replicas,
		volume:         nil,
		volumeMount:    nil,
	}

	return c.ensureStatefulSet(proxysql, proxysql.Spec.UpdateStrategy, opts)
}

func (c *Controller) checkStatefulSet(proxysql *api.ProxySQL, stsName string) error {
	// StatefulSet for ProxySQL database
	statefulSet, err := c.Client.AppsV1().StatefulSets(proxysql.Namespace).Get(stsName, metav1.GetOptions{})
	if err != nil {
		if kerr.IsNotFound(err) {
			return nil
		}
		return err
	}

	if statefulSet.Labels[api.LabelDatabaseKind] != api.ResourceKindProxySQL ||
		statefulSet.Labels[api.LabelProxySQLName] != proxysql.Name {
		return fmt.Errorf(`intended statefulSet "%v/%v" already exists`, proxysql.Namespace, stsName)
	}

	return nil
}

func upsertCustomConfig(template core.PodTemplateSpec, configSource *core.VolumeSource) core.PodTemplateSpec {
	for i, container := range template.Spec.Containers {
		if container.Name == api.ResourceSingularProxySQL {
			configVolumeMount := core.VolumeMount{
				Name:      "custom-config",
				MountPath: api.ProxySQLCustomConfigMountPath,
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, configVolumeMount)
			template.Spec.Containers[i].VolumeMounts = volumeMounts

			configVolume := core.Volume{
				Name:         "custom-config",
				VolumeSource: *configSource,
			}

			volumes := template.Spec.Volumes
			volumes = core_util.UpsertVolume(volumes, configVolume)
			template.Spec.Volumes = volumes
			break
		}
	}

	return template
}

func (c *Controller) ensureStatefulSet(
	proxysql *api.ProxySQL,
	updateStrategy apps.StatefulSetUpdateStrategy,
	opts workloadOptions) (kutil.VerbType, error) {
	// Take value of podTemplate
	var pt ofst.PodTemplateSpec
	if opts.podTemplate != nil {
		pt = *opts.podTemplate
	}
	if err := c.checkStatefulSet(proxysql, opts.stsName); err != nil {
		return kutil.VerbUnchanged, err
	}

	// Create statefulSet for ProxySQL database
	statefulSetMeta := metav1.ObjectMeta{
		Name:      opts.stsName,
		Namespace: proxysql.Namespace,
	}

	ref, rerr := reference.GetReference(clientsetscheme.Scheme, proxysql)
	if rerr != nil {
		return kutil.VerbUnchanged, rerr
	}

	readinessProbe := pt.Spec.ReadinessProbe
	if readinessProbe != nil && structs.IsZero(*readinessProbe) {
		readinessProbe = nil
	}
	livenessProbe := pt.Spec.LivenessProbe
	if livenessProbe != nil && structs.IsZero(*livenessProbe) {
		livenessProbe = nil
	}

	statefulSet, vt, err := app_util.CreateOrPatchStatefulSet(c.Client, statefulSetMeta, func(in *apps.StatefulSet) *apps.StatefulSet {
		in.Labels = opts.labels
		in.Annotations = pt.Controller.Annotations
		core_util.EnsureOwnerReference(&in.ObjectMeta, ref)

		in.Spec.Replicas = opts.replicas
		in.Spec.ServiceName = opts.gvrSvcName
		in.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: opts.selectors,
		}
		in.Spec.Template.Labels = opts.selectors
		in.Spec.Template.Annotations = pt.Annotations
		in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
			in.Spec.Template.Spec.InitContainers,
			pt.Spec.InitContainers,
		)
		in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
			in.Spec.Template.Spec.Containers,
			core.Container{
				Name:            opts.conatainerName,
				Image:           opts.image,
				ImagePullPolicy: core.PullIfNotPresent,
				Command:         opts.cmd,
				Args:            opts.args,
				Ports:           opts.ports,
				Env:             core_util.UpsertEnvVars(opts.envList, pt.Spec.Env...),
				Resources:       pt.Spec.Resources,
				Lifecycle:       pt.Spec.Lifecycle,
				LivenessProbe:   livenessProbe,
				ReadinessProbe:  readinessProbe,
				VolumeMounts:    opts.volumeMount,
			})

		in.Spec.Template.Spec.InitContainers = core_util.UpsertContainers(
			in.Spec.Template.Spec.InitContainers,
			opts.initContainers,
		)

		if opts.monitorContainer != nil && proxysql.GetMonitoringVendor() == mona.VendorPrometheus {
			in.Spec.Template.Spec.Containers = core_util.UpsertContainer(
				in.Spec.Template.Spec.Containers, *opts.monitorContainer)
		}

		in.Spec.Template.Spec.Volumes = core_util.UpsertVolume(in.Spec.Template.Spec.Volumes, opts.volume...)

		//in = upsertEnv(in, proxysql)
		in = upsertDataVolume(in, proxysql)

		if opts.configSource != nil {
			in.Spec.Template = upsertCustomConfig(in.Spec.Template, opts.configSource)
		}

		in.Spec.Template.Spec.NodeSelector = pt.Spec.NodeSelector
		in.Spec.Template.Spec.Affinity = pt.Spec.Affinity
		if pt.Spec.SchedulerName != "" {
			in.Spec.Template.Spec.SchedulerName = pt.Spec.SchedulerName
		}
		in.Spec.Template.Spec.Tolerations = pt.Spec.Tolerations
		in.Spec.Template.Spec.ImagePullSecrets = pt.Spec.ImagePullSecrets
		in.Spec.Template.Spec.PriorityClassName = pt.Spec.PriorityClassName
		in.Spec.Template.Spec.Priority = pt.Spec.Priority
		in.Spec.Template.Spec.SecurityContext = pt.Spec.SecurityContext

		if c.EnableRBAC {
			in.Spec.Template.Spec.ServiceAccountName = pt.Spec.ServiceAccountName
		}

		in.Spec.UpdateStrategy = updateStrategy
		return in
	})

	if err != nil {
		return kutil.VerbUnchanged, err
	}

	// Check StatefulSet Pod status
	if vt != kutil.VerbUnchanged {
		if err := c.checkStatefulSetPodStatus(statefulSet); err != nil {
			return kutil.VerbUnchanged, err
		}
		c.recorder.Eventf(
			proxysql,
			core.EventTypeNormal,
			eventer.EventReasonSuccessful,
			"Successfully %v StatefulSet %v/%v",
			vt, proxysql.Namespace, opts.stsName,
		)
	}

	return vt, nil
}

func upsertDataVolume(statefulSet *apps.StatefulSet, proxysql *api.ProxySQL) *apps.StatefulSet {
	for i, container := range statefulSet.Spec.Template.Spec.Containers {
		if container.Name == api.ResourceSingularProxySQL {
			volumeMount := core.VolumeMount{
				Name:      "data",
				MountPath: api.ProxySQLDataMountPath,
			}
			volumeMounts := container.VolumeMounts
			volumeMounts = core_util.UpsertVolumeMount(volumeMounts, volumeMount)
			statefulSet.Spec.Template.Spec.Containers[i].VolumeMounts = volumeMounts

			pvcSpec := proxysql.Spec.Storage
			if proxysql.Spec.StorageType == api.StorageTypeEphemeral {
				ed := core.EmptyDirVolumeSource{}
				if pvcSpec != nil {
					if sz, found := pvcSpec.Resources.Requests[core.ResourceStorage]; found {
						ed.SizeLimit = &sz
					}
				}
				statefulSet.Spec.Template.Spec.Volumes = core_util.UpsertVolume(
					statefulSet.Spec.Template.Spec.Volumes,
					core.Volume{
						Name: "data",
						VolumeSource: core.VolumeSource{
							EmptyDir: &ed,
						},
					})
			} else {
				if len(pvcSpec.AccessModes) == 0 {
					pvcSpec.AccessModes = []core.PersistentVolumeAccessMode{
						core.ReadWriteOnce,
					}
					log.Infof(`Using "%v" as AccessModes in .spec.storage`, core.ReadWriteOnce)
				}

				claim := core.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: *pvcSpec,
				}
				if pvcSpec.StorageClassName != nil {
					claim.Annotations = map[string]string{
						"volume.beta.kubernetes.io/storage-class": *pvcSpec.StorageClassName,
					}
				}
				statefulSet.Spec.VolumeClaimTemplates = core_util.UpsertVolumeClaim(statefulSet.Spec.VolumeClaimTemplates, claim)
			}
			break
		}
	}
	return statefulSet
}

//
//// upsertUserEnv add/overwrite env from user provided env in crd spec
//func upsertEnv(statefulSet *apps.StatefulSet, proxysql *api.ProxySQL) *apps.StatefulSet {
//	for i, container := range statefulSet.Spec.Template.Spec.Containers {
//		if container.Name == api.ResourceSingularProxySQL || container.Name == "exporter" {
//			envs := []core.EnvVar{
//				{
//					Name: "MYSQL_ROOT_PASSWORD",
//					ValueFrom: &core.EnvVarSource{
//						SecretKeyRef: &core.SecretKeySelector{
//							LocalObjectReference: core.LocalObjectReference{
//								Name: proxysql.Spec.DatabaseSecret.SecretName,
//							},
//							Key: MySQLPasswordKey,
//						},
//					},
//				},
//				{
//					Name: "MYSQL_ROOT_USERNAME",
//					ValueFrom: &core.EnvVarSource{
//						SecretKeyRef: &core.SecretKeySelector{
//							LocalObjectReference: core.LocalObjectReference{
//								Name: proxysql.Spec.DatabaseSecret.SecretName,
//							},
//							Key: mysqlUserKey,
//						},
//					},
//				},
//			}
//
//			statefulSet.Spec.Template.Spec.Containers[i].Env = core_util.UpsertEnvVars(container.Env, envs...)
//		}
//	}
//
//	return statefulSet
//}

func (c *Controller) checkStatefulSetPodStatus(statefulSet *apps.StatefulSet) error {
	err := core_util.WaitUntilPodRunningBySelector(
		c.Client,
		statefulSet.Namespace,
		statefulSet.Spec.Selector,
		int(types.Int32(statefulSet.Spec.Replicas)),
	)
	if err != nil {
		return err
	}
	return nil
}
