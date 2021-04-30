module github.com/linki/encrypted-secrets

go 1.15

require (
	cloud.google.com/go v0.51.0
	github.com/aws/aws-sdk-go-v2 v1.3.4
	github.com/aws/aws-sdk-go-v2/config v1.1.6
	github.com/aws/aws-sdk-go-v2/service/kms v1.2.2
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.2.2
	github.com/go-logr/logr v0.3.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/stretchr/testify v1.5.1
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	sigs.k8s.io/controller-runtime v0.7.2
)
