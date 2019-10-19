package controller

import (
	"github.com/appscode/go/log"
	core_util "kmodules.xyz/client-go/core/v1"
	"kmodules.xyz/client-go/tools/queue"
	"kubedb.dev/apimachinery/apis"
	api "kubedb.dev/apimachinery/apis/kubedb/v1alpha1"
	"kubedb.dev/apimachinery/client/clientset/versioned/typed/kubedb/v1alpha1/util"
)

func (c *Controller) initWatcher() {
	c.proxysqlInformer = c.KubedbInformerFactory.Kubedb().V1alpha1().ProxySQLs().Informer()
	c.proxysqlQueue = queue.New("ProxySQL", c.MaxNumRequeues, c.NumThreads, c.runProxySQL)
	c.proxysqlLister = c.KubedbInformerFactory.Kubedb().V1alpha1().ProxySQLs().Lister()
	c.proxysqlInformer.AddEventHandler(queue.NewObservableUpdateHandler(c.proxysqlQueue.GetQueue(), apis.EnableStatusSubresource))
}

func (c *Controller) runProxySQL(key string) error {
	log.Debugln("started processing, key:", key)
	obj, exists, err := c.proxysqlInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("ProxySQL %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a ProxySQL was recreated with the same name
		proxysql := obj.(*api.ProxySQL).DeepCopy()
		if proxysql.DeletionTimestamp != nil {
			if core_util.HasFinalizer(proxysql.ObjectMeta, api.GenericKey) {
				if err := c.terminate(proxysql); err != nil {
					log.Errorln(err)
					return err
				}
				proxysql, _, err = util.PatchProxySQL(c.ExtClient.KubedbV1alpha1(), proxysql, func(in *api.ProxySQL) *api.ProxySQL {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				})
				return err
			}
		} else {
			proxysql, _, err = util.PatchProxySQL(c.ExtClient.KubedbV1alpha1(), proxysql, func(in *api.ProxySQL) *api.ProxySQL {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			})
			if err != nil {
				return err
			}
			if err := c.create(proxysql); err != nil {
				log.Errorln(err)
				c.pushFailureEvent(proxysql, err.Error())
				return err
			}
		}
	}
	return nil
}
