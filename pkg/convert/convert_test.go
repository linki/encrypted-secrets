package convert

import (
	"testing"

	"github.com/stretchr/testify/suite"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8sv1alpha1 "github.com/linki/encrypted-secrets/pkg/apis/k8s/v1alpha1"
)

type ConvertSuite struct {
	suite.Suite
}

func (suite *ConvertSuite) TestConvert() {
	for _, tc := range []struct {
		given    corev1.Secret
		expected k8sv1alpha1.EncryptedSecret
	}{
		// TODO
		{
			given: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
					Labels: map[string]string{
						"key": "value",
					},
				},
				Data: map[string][]byte{
					"key": []byte("value"),
				},
			},
			expected: k8sv1alpha1.EncryptedSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
					Labels: map[string]string{
						"key": "value",
					},
				},
				Spec: k8sv1alpha1.EncryptedSecretSpec{
					Data: map[string][]byte{
						"key": []byte("value"),
					},
				},
			},
		},
	} {
		converted := convert(tc.given)
		suite.Equal(tc.expected, converted)
	}
}

func TestConvertSuite(t *testing.T) {
	suite.Run(t, new(ConvertSuite))
}
