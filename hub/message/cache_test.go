package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacher_NewCacher_local(t *testing.T) {
	//Arrange/Act
	sut := NewCacher(false, nil)

	//Assert
	assert.NotNil(t, sut)
	_, ok := sut.(*localCache)

	assert.True(t, ok)
}

func TestCacher_NewCacher_distributed(t *testing.T) {
	//Arrange/Act
	sut := NewCacher(true, nil)

	//Assert
	assert.Nil(t, sut)
}
