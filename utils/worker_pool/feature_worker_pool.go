package workerpool

import (
	"github.com/bldsoft/gost/config/feature"
)

type FeatureWorkerPool struct {
	WorkerPool
	workerNFeature *feature.Int
	defaultValue   int // used when feature value <= 0
}

func NewWorkerPoolWithFeature(workerN *feature.Int, defaultValue int) *FeatureWorkerPool {
	wp := &FeatureWorkerPool{
		workerNFeature: workerN,
	}
	wp.SetDefaultValue(defaultValue)
	workerN.AddOnChangeHandler(wp.setWorkerN)
	return wp
}

func (wp *FeatureWorkerPool) setWorkerN(n int) {
	if n <= 0 {
		n = wp.defaultValue
	}
	n = max(1, n)
	wp.WorkerPool.SetWorkerN(int64(n))
}

func (wp *FeatureWorkerPool) SetDefaultValue(n int) *FeatureWorkerPool {
	wp.defaultValue = n
	wp.setWorkerN(wp.workerNFeature.Get())
	return wp
}
