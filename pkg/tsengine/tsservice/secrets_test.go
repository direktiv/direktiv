package tsservice_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/tsengine/tsservice"
	"github.com/stretchr/testify/assert"
)

func TestSecretsMultiple(t *testing.T) {

	def := `
	const sec1 = getSecret({
		name: "mysecret1"
	})

	function start() {
		const sec1 = getSecret({
			name: "mysecret2"
		})

		const sec1 = getSecret({
			name: "mysecret3"
		})
	
	}
	`

	c, _ := tsservice.New("", def)
	fi, _ := c.CompileFlow()
	assert.Len(t, fi.Secrets, 3)
}
