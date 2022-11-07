package utils

import (
	"bytes"
	"text/template"

	corev1 "k8s.io/api/core/v1"
)

type MySQLPassword struct {
	superUserPassword string
}

var mycnfTmpl = template.Must(template.New("my.cnf").Parse(`[client]
user={{printf "%q" .User}}
password={{printf "%q" .Password}}
{{if .Socket -}}
socket={{printf "%q" .Socket}}
{{end}}`))

func FormatMyClientCnf(user, pwd, socket string) []byte {
	buf := new(bytes.Buffer)
	err := mycnfTmpl.Execute(buf, struct {
		User     string
		Password string
		Socket   string
	}{
		user,
		pwd,
		socket,
	})
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func (p MySQLPassword) ToMyClientCnfSecret() *corev1.Secret {
	return &corev1.Secret{
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			RootUserClientConf: FormatMyClientCnf(RootUser, p.superUserPassword, ""),
		},
	}
}

func NewMyclientCnfFromSecret(s *corev1.Secret) *MySQLPassword {
	return &MySQLPassword{
		superUserPassword: string(s.Data[RootUserPasswordKey]),
	}
}
