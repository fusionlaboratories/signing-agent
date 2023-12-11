package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacher_NewCacher_local(t *testing.T) {
	//Arrange/Act
	sut := NewCacher(false, nil, nil)

	//Assert
	assert.NotNil(t, sut)
	_, ok := sut.(*localCache)

	assert.True(t, ok)
}

func TestCacher_NewCacher_distributed(t *testing.T) {
	//Arrange/Act
	sut := NewCacher(true, nil, nil)

	// Assert
	assert.NotNil(t, sut)
	_, ok := sut.(*distributedCache)

	assert.True(t, ok)
}
