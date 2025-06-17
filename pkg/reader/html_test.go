package reader_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tim8842/tender-data-loader/pkg/reader"
)

func TestReadHtmlFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("../web/test", "read.html")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := "<html><body>Test</body></html>"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	result := reader.ReadHtmlFile(tmpFile.Name())
	require.Equal(t, []byte(content), result)
}
