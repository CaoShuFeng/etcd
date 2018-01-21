package cmd

import "math/rand"
import "time"

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func getRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGIHJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func getRandomStringRandomLen(min, max int) string {
	ran := max - min + 1
	length := min + r.Intn(ran)
	return getRandomString(length)
}

func getAuditKey() string {
	pesudoNameSpace := getRandomStringRandomLen(7, 16)
	pesudoAuditName := getRandomStringRandomLen(6, 15)
	return "/registry/Audit/" + pesudoNameSpace + "/" + pesudoAuditName
}

func getAuditJson() string {
	audit_json_part1 := `{kind:Event,apiVersion:audit.k8s.io/v1beta1,metadata:{creationTimestamp:2018-01-02T02:46:14Z},level:Metadata,timestamp:2018-01-02T02:46:14Z,auditID:`
	audit_json_part2 := `,"kind":"Event","apiVersion":"audit.k8s.io/v1beta1","metadata":{"creationTimestamp":"2018-01-19T18:31:48Z"},"level":"RequestResponse","timestamp":"2018-01-19T18:31:48Z","auditID":"3a41e149-065d-49eb-9e86-aec90ca03117","stage":"ResponseComplete","requestURI":"/apis/authorization.k8s.io/v1","verb":"get","user":{"username":"system:unsecured","groups":["system:masters","system:authenticated"]},"sourceIPs":["172.16.29.130"],"responseStatus":{"metadata":{},"code":200},"responseObject":{"kind":"APIResourceList","apiVersion":"v1","groupVersion":"authorization.k8s.io/v1","resources":[{"name":"localsubjectaccessreviews","singularName":"","namespaced":true,"kind":"LocalSubjectAccessReview","verbs":["create"]},{"name":"selfsubjectaccessreviews","singularName":"","namespaced":false,"kind":"SelfSubjectAccessReview","verbs":["create"]},{"name":"selfsubjectrulesreviews","singularName":"","namespaced":false}}}`
	pesudoUID := getRandomStringRandomLen(1, 155)
	return audit_json_part1 + pesudoUID + audit_json_part2
}
