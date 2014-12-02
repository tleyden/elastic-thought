package caffe

func (l LayerParameter) IsTrainingPhase() bool {

	includes := l.Include
	for _, include := range includes {
		if *include.Phase == Phase_TRAIN {
			return true
		}
	}

	return false
}

func (l LayerParameter) IsTestingPhase() bool {
	includes := l.Include
	for _, include := range includes {
		if *include.Phase == Phase_TEST {
			return true
		}
	}

	return false

}

func (l LayerParameter) GetImageDataSource() string {
	return *l.ImageDataParam.Source
}
