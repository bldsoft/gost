package routing

func init() {
	RegisterCondition[*FieldCondition]("generic")

	RegisterValueExtractor[HostExtractor]("host")

	RegisterValueMatcher[MatcherAnyOf]("anyOf")

	RegisterAction[ActionRedirect]("redirect")
}

var actionMarshaller = &PolymorphMarshaller[Action]{}

func RegisterAction[T Action](name string) {
	var value T
	actionMarshaller.Register(name, value)
}

var conditionPolymorphMarshaller = &PolymorphMarshaller[Condition]{}

func RegisterCondition[T Condition](name string) {
	var value T
	conditionPolymorphMarshaller.Register(name, value)
}

var valueExtractorMarshaller = &PolymorphMarshaller[ValueExtractor]{}

func RegisterValueExtractor[T ValueExtractor](name string) {
	var value T
	valueExtractorMarshaller.Register(name, value)
}

var valueMatcherMarshaller = &PolymorphMarshaller[ValueMatcher]{}

func RegisterValueMatcher[T ValueMatcher](name string) {
	var value T
	valueMatcherMarshaller.Register(name, value)
}
