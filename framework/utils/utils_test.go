package utils_test

import (
	"encoder/framework/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsJson(t *testing.T) {
	json := `{"id": "dsdsf", "file_path": "dfdf", "status": "sdfdsf"}`
	err := utils.IsJson(json)
	require.Nil(t, err)

	json = `-{"id": "dsdsf", "file_path": "dfdf", "status": "sdfdsf"}`
	err = utils.IsJson(json)
	require.Error(t, err)
}
