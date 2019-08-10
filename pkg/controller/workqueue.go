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
	c.pxInformer = c.KubedbInformerFactory.Kubedb().V1alpha1().PerconaXtraDBs().Informer()
	c.pxQueue = queue.New("PerconaXtraDB", c.MaxNumRequeues, c.NumThreads, c.runPerconaXtraDB)
	c.pxLister = c.KubedbInformerFactory.Kubedb().V1alpha1().PerconaXtraDBs().Lister()
	c.pxInformer.AddEventHandler(queue.NewObservableUpdateHandler(c.pxQueue.GetQueue(), apis.EnableStatusSubresource))
}

func (c *Controller) runPerconaXtraDB(key string) error {
	log.Debugln("started processing, key:", key)
	obj, exists, err := c.pxInformer.GetIndexer().GetByKey(key)
	if err != nil {
		log.Errorf("Fetching object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		log.Debugf("PerconaXtraDB %s does not exist anymore", key)
	} else {
		// Note that you also have to check the uid if you have a local controlled resource, which
		// is dependent on the actual instance, to detect that a PerconaXtraDB was recreated with the same name
		px := obj.(*api.PerconaXtraDB).DeepCopy()
		if px.DeletionTimestamp != nil {
			if core_util.HasFinalizer(px.ObjectMeta, api.GenericKey) {
				if err := c.terminate(px); err != nil {
					log.Errorln(err)
					return err
				}
				px, _, err = util.PatchPerconaXtraDB(c.ExtClient.KubedbV1alpha1(), px, func(in *api.PerconaXtraDB) *api.PerconaXtraDB {
					in.ObjectMeta = core_util.RemoveFinalizer(in.ObjectMeta, api.GenericKey)
					return in
				})
				return err
			}
		} else {
			px, _, err = util.PatchPerconaXtraDB(c.ExtClient.KubedbV1alpha1(), px, func(in *api.PerconaXtraDB) *api.PerconaXtraDB {
				in.ObjectMeta = core_util.AddFinalizer(in.ObjectMeta, api.GenericKey)
				return in
			})
			if err != nil {
				return err
			}
			if err := c.create(px); err != nil {
				log.Errorln(err)
				c.pushFailureEvent(px, err.Error())
				return err
			}
		}
	}
	return nil
}
