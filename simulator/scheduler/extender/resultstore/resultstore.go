package resultstore

import (
	"encoding/json"
	"sync"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/extender/annotation"
)

type Store interface {
	AddStoredResultToPod(pod *v1.Pod)
	DeleteData(pod v1.Pod)
	AddFilterResult(args extenderv1.ExtenderArgs, result extenderv1.ExtenderFilterResult, hostName string)
	AddPrioritizeResult(args extenderv1.ExtenderArgs, result extenderv1.HostPriorityList, hostName string)
	AddPreemptResult(args extenderv1.ExtenderPreemptionArgs, result extenderv1.ExtenderPreemptionResult, hostName string)
	AddBindResult(args extenderv1.ExtenderBindingArgs, result extenderv1.ExtenderBindingResult, hostName string)
}

// store has results of all extenders.
// It manages all extenders results.
type store struct {
	mu *sync.Mutex

	results map[key]*result
	weight  map[string]int32
}

// key is the key of result map on Store.
// key is created from namespace and podName.
type key string

type result struct {
	filter map[string]extenderv1.ExtenderFilterResult

	prioritize map[string]extenderv1.HostPriorityList

	preempt map[string]extenderv1.ExtenderPreemptionResult

	bind map[string]extenderv1.ExtenderBindingResult
}

func New() Store {
	s := &store{
		mu:      new(sync.Mutex),
		results: map[key]*result{},
		weight:  map[string]int32{}, // TODO: Not needed?
	}
	return s
}

// newKey creates key with namespace and podName.
func newKey(namespace, podName string) key {
	k := namespace + "/" + podName
	return key(k)
}

func newData() *result {
	return &result{
		filter:     map[string]extenderv1.ExtenderFilterResult{},
		prioritize: map[string]extenderv1.HostPriorityList{},
		preempt:    map[string]extenderv1.ExtenderPreemptionResult{},
		bind:       map[string]extenderv1.ExtenderBindingResult{},
	}
}

// AddStoredResultToPod adds all data corresponding to the specified key to the pod.
func (s *store) AddStoredResultToPod(pod *v1.Pod) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(pod.Namespace, pod.Name)
	if _, ok := s.results[k]; !ok {
		// Store doesn't have any scheduling result of the Pod.
		return
	}

	if err := s.addFilterResultToPod(pod); err != nil {
		klog.Errorf("failed to add filtering result to the pod: %+v", err)
		return
	}

	if err := s.addPrioritizeResultToPod(pod); err != nil {
		klog.Errorf("failed to add prioritize result to the pod: %+v", err)
		return
	}

	if err := s.addPreemptResultToPod(pod); err != nil {
		klog.Errorf("failed to add preempt result to the pod: %+v", err)
		return
	}

	if err := s.addBindResultToPod(pod); err != nil {
		klog.Errorf("failed to add bind result to the pod: $+v", err)
		return
	}
}

func (s *store) addFilterResultToPod(pod *v1.Pod) error {
	k := newKey(pod.Namespace, pod.Name)
	results, err := json.Marshal(s.results[k].filter)
	if err != nil {
		return xerrors.Errorf("encode Filter results to json: %w", err)
	}
	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.ExtenderFilterResultAnnotationKey, string(results))
	return nil
}

func (s *store) addPrioritizeResultToPod(pod *v1.Pod) error {
	k := newKey(pod.Namespace, pod.Name)
	results, err := json.Marshal(s.results[k].prioritize)
	if err != nil {
		return xerrors.Errorf("encode Prioritize results to json: %w", err)
	}
	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.ExtenderPrioritizeResultAnnotationKey, string(results))
	return nil
}

func (s *store) addPreemptResultToPod(pod *v1.Pod) error {
	k := newKey(pod.Namespace, pod.Name)
	results, err := json.Marshal(s.results[k].preempt)
	if err != nil {
		return xerrors.Errorf("encode Preempt results to json: %w", err)
	}
	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.ExtenderPreemptResultAnnotationKey, string(results))
	return nil
}

func (s *store) addBindResultToPod(pod *v1.Pod) error {
	k := newKey(pod.Namespace, pod.Name)
	results, err := json.Marshal(s.results[k].bind)
	if err != nil {
		return xerrors.Errorf("encode Bind results to json: %w", err)
	}
	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.ExtenderBindResultAnnotationKey, string(results))
	return nil
}

// AddFilterResult stores the filtering result.
func (s *store) AddFilterResult(args extenderv1.ExtenderArgs, result extenderv1.ExtenderFilterResult, hostName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(args.Pod.Namespace, args.Pod.Name)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}
	s.results[k].filter[hostName] = result
}

// AddPrioritizeResult stores the prioritizing result.
func (s *store) AddPrioritizeResult(args extenderv1.ExtenderArgs, result extenderv1.HostPriorityList, hostName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(args.Pod.Namespace, args.Pod.Name)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}
	s.results[k].prioritize[hostName] = result
}

// AddPreemptResult stores the preempting result.
func (s *store) AddPreemptResult(args extenderv1.ExtenderPreemptionArgs, result extenderv1.ExtenderPreemptionResult, hostName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(args.Pod.Namespace, args.Pod.Name)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}
	s.results[k].preempt[hostName] = result
}

// AddBindResult stores the binding result.
func (s *store) AddBindResult(args extenderv1.ExtenderBindingArgs, result extenderv1.ExtenderBindingResult, hostName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(args.PodNamespace, args.PodName)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}
	s.results[k].bind[hostName] = result
}

// DeleteData deletes the data corresponding to the specified Pod.
func (s *store) DeleteData(pod v1.Pod) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleteData(newKey(pod.Namespace, pod.Name))
}

// deleteData deletes the result stored with the given key.
func (s *store) deleteData(k key) {
	delete(s.results, k)
}
