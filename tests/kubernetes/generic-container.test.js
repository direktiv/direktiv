import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import { fileURLToPath } from 'url'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'

const testNamespace = basename(fileURLToPath(import.meta.url))

const genericContainerWorkflow = `
direktiv_api: workflow/v1
functions:
- id: test
  image: alpine
  type: knative-workflow
  cmd: /usr/share/direktiv/direktiv-cmd
states:
- id: test
  type: action
  action:
    function: test
    input: 
      data:
        commands:
        - command: echo -n data
`

describe('Test generic container', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, testNamespace)

	helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'wf1.yaml', 'workflow',
		genericContainerWorkflow,
	)

	helpers.sleep(100)

	it(`should invoke workflow`, async () => {
		const res = await request(config.getDirektivBaseUrl()).post(`/api/v2/namespaces/${ testNamespace }/instances?path=wf1.yaml&wait=true`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.return[0].Output).toEqual('data')
	})
})
