package lastapplied

import (
	"archive/zip"
	"bytes"
	"encoding/base64"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	json "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
)

func GetOriginalSts(sts *appsv1.StatefulSet) (appsv1.StatefulSet, error) {
	stsOrig, err := patch.DefaultAnnotator.GetOriginalConfiguration(sts)
	stsOrigObj := appsv1.StatefulSet{}
	err = json.ConfigCompatibleWithStandardLibrary.Unmarshal(stsOrig, &stsOrigObj)
	if err != nil {
		logrus.Debug("cannot deserialize stsOrig")
	}
	return stsOrigObj, err
}

func EncodeLastAppliedConfigAnnotation(originalStatefulSet appsv1.StatefulSet) (string, error) {
	marshal, err := json.ConfigCompatibleWithStandardLibrary.Marshal(originalStatefulSet)
	if err != nil {
		return "", err
	}
	marshalWithoutNulls, _, err := patch.DeleteNullInJson(marshal)
	if err != nil {
		return "", err
	}
	zipped, err := zipAndBase64EncodeAnnotation(marshalWithoutNulls)
	if err != nil {
		return "", err
	}
	return zipped, nil
}

func zipAndBase64EncodeAnnotation(original []byte) (string, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	writer, err := zipWriter.Create("original")
	if err != nil {
		return "", err
	}
	_, err = writer.Write(original)
	if err != nil {
		return "", err
	}
	err = zipWriter.Close()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
