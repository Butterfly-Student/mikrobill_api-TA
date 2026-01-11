package testing

type TestingDomain interface{}

type testingDomain struct{}

func NewTestingDomain() TestingDomain {
	return &testingDomain{}
}
