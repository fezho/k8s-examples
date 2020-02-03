package controller

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	batchinformers "k8s.io/client-go/informers/batch/v1"
	clientset "k8s.io/client-go/kubernetes"
	batchv1listers "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

type JobController struct {
	*Controller

	retentionInSeconds int64

	kubeClient clientset.Interface
	jobLister  batchv1listers.JobLister
}

func NewJobController(kubeClient clientset.Interface, jobInformer batchinformers.JobInformer, retentionInSeconds int64) *JobController {
	jc := &JobController{
		kubeClient:         kubeClient,
		jobLister:          jobInformer.Lister(),
		retentionInSeconds: retentionInSeconds,
	}

	jc.Controller = NewController("job-retention", jc.handleJob, jobInformer.Informer().HasSynced)

	jobInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			jc.enqueue(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			jc.enqueue(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			klog.Infof("ignore deleted job")
		},
	})

	return jc
}

func (c *JobController) enqueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}

	c.queue.Add(key)
}

func (c *JobController) handleJob(namespace, name string) error {
	job, err := c.jobLister.Jobs(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("job '%s/%s' in work queue no longer exists", namespace, name))
			return nil
		}
		return err
	}

	// skip job is not completed
	if job.Status.CompletionTime == nil {
		klog.Infof("job %s/%s is not completed", namespace, name)
		return nil
	}

	// retention logic
	if time.Now().Unix()-job.Status.CompletionTime.Time.Unix() > c.retentionInSeconds {
		err = c.deleteJobAndPods(namespace, name)
		if err != nil {
			klog.Errorf("failed to delete job %s/%s or its corresponding pods, error: %v", namespace, name, err)
			return err
		} else {
			klog.Infof("delete exceeded retention job %s/%s", namespace, name)
		}
	}

	return nil
}

func (c *JobController) deleteJobAndPods(namespace, name string) error {
	podsSelector := fmt.Sprintf("job-name=%s", name)
	listOpts := metav1.ListOptions{LabelSelector: podsSelector}
	err := c.kubeClient.CoreV1().Pods(namespace).DeleteCollection(&v1.DeleteOptions{}, listOpts)
	if err != nil {
		return err
	}
	return c.kubeClient.BatchV1().Jobs(namespace).Delete(name, &v1.DeleteOptions{})
}
