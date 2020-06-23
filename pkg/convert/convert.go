package convert

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8sv1alpha1 "github.com/linki/encrypted-secrets/pkg/apis/k8s/v1alpha1"
)

func convert(secret corev1.Secret) k8sv1alpha1.EncryptedSecret {
	return k8sv1alpha1.EncryptedSecret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secret.Name,
			Namespace: secret.Namespace,
			Labels:    secret.Labels,
		},
		Spec: k8sv1alpha1.EncryptedSecretSpec{
			Data: map[string][]byte{
				"key": []byte("value"),
			},
		},
	}
}
