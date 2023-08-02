package util

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/require"

	"github.com/qredo/signing-agent/config"
)

func Test_StoreFactory_CreateStore_Creates_file_store(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Store: config.Store{
			Type: "file",
		},
	}

	// Act
	sut := CreateStore(cfg)

	// Assert
	require.NotNil(t, sut)
	assert.Equal(t, reflect.TypeOf(sut).String(), "*util.FileStore")
}

func Test_StoreFactory_CreateStore_Creates_oci_store(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Store: config.Store{
			Type: "oci",
		},
	}

	// Act
	sut := CreateStore(cfg)

	// Assert
	require.NotNil(t, sut)
	assert.Equal(t, reflect.TypeOf(sut).String(), "*util.OciStore")
}

func Test_StoreFactory_CreateStore_Creates_aws_store(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Store: config.Store{
			Type: "aws",
		},
	}

	// Act
	sut := CreateStore(cfg)

	// Assert
	require.NotNil(t, sut)
	assert.Equal(t, reflect.TypeOf(sut).String(), "*util.AWSStore")
}

func Test_StoreFactory_CreateStore_Creates_gcp_store(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		Store: config.Store{
			Type: "gcp",
		},
	}

	// Act
	sut := CreateStore(cfg)

	// Assert
	require.NotNil(t, sut)
	assert.Equal(t, reflect.TypeOf(sut).String(), "*util.GCPStore")
}
