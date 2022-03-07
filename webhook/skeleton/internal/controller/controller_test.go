package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.mgmt.arms-dev.net/go-common/logger"
)

type mockService struct {
	useAnnotation bool
}

func (m *mockService) UseAnnotations() (bool, error) {
	return m.useAnnotation, nil
}

func TestGetIngressClassName(t *testing.T) {

	testCases := []struct {
		name            string
		useAnnotation   bool
		annotationValue string
		specValue       string
		expected        string
	}{
		{
			name:          "empty-annotation",
			useAnnotation: true,
			expected:      defaultIngressClassAnnotationValue,
		},
		{
			name:          "empty-spec",
			useAnnotation: false,
			expected:      defaultIngressClassName,
		},
		{
			name:            "default-annotation-use-spec",
			useAnnotation:   false,
			annotationValue: defaultIngressClassAnnotationValue,
			expected:        defaultIngressClassName,
		},
		{
			name:          "default-spec-use-annotation",
			useAnnotation: true,
			specValue:     defaultIngressClassName,
			expected:      defaultIngressClassAnnotationValue,
		},
		{
			name:            "defaults-use-annotation",
			useAnnotation:   true,
			annotationValue: defaultIngressClassAnnotationValue,
			specValue:       defaultIngressClassName,
			expected:        defaultIngressClassAnnotationValue,
		},
		{
			name:            "defaults-use-spec",
			useAnnotation:   false,
			annotationValue: defaultIngressClassAnnotationValue,
			specValue:       defaultIngressClassName,
			expected:        defaultIngressClassName,
		},
		{
			name:          "nondefault-spec-use-annotation",
			useAnnotation: true,
			specValue:     "foo",
			expected:      "foo",
		},
		{
			name:            "difference-use-annotation",
			useAnnotation:   true,
			annotationValue: "bar",
			specValue:       "foo",
			expected:        "foo",
		},
		{
			name:            "difference-use-spec",
			useAnnotation:   false,
			annotationValue: "bar",
			specValue:       "foo",
			expected:        "foo",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			s := &service{
				log: logger.NewEntry(),
				ingressService: &mockService{
					useAnnotation: testCase.useAnnotation,
				},
			}

			assert.Equal(t, testCase.expected, s.getIngressClassName(map[string]string{
				ingressClassAnnotationKey: testCase.annotationValue,
			}, &testCase.specValue, testCase.useAnnotation))
		})
	}
}
