package tsservice_test

import (
	"testing"

	"github.com/direktiv/direktiv/pkg/tsengine/tsservice"
	"github.com/stretchr/testify/assert"
)

func TestFilesMultiple(t *testing.T) {

	def := `
	const fileOne = getFile({
		name: "first.txt",
		permission: 755,
		scope: "shared",
	});	  

	function start() {

		var fileThree = getFile({
			name: "third.txt",
			permission: 755,
			scope: "shared",
		});	 

		var fileTwo = getFile({
			name: "second.txt",
			permission: 755,
			scope: "shared",
		});	  
	}
	`

	c, _ := tsservice.New("", def)
	fi, _ := c.CompileFlow()
	assert.Len(t, fi.Files, 3)
}
