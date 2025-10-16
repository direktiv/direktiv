import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10 } from '../common/retry'

const testNamespace = 'patches'

const genericContainerWorkflow = `
direktiv_api: workflow/v1
functions:
- id: test
  image: alpine
  type: knative-workflow
  cmd: /usr/share/direktiv/direktiv-cmd
  patches:
  - op: add
    path: /spec/template/spec/containers/0/env/-
    Value: { "name": "MYENV", "value": "value"}
states:
- id: test 
  type: action
  action:
    function: test
    input: 
      data:
        commands:
        - command: echo -n $MYENV
`

describe('Test generic container', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/',
		'wf1.yaml',
		'workflow',
		genericContainerWorkflow,
	)

	retry10(`should invoke workflow`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).post(
			`/api/v2/namespaces/${testNamespace}/instances?path=wf1.yaml&wait=true`,
		)
		expect(res.statusCode).toEqual(200)
		expect(res.body.return[0].Output).toEqual('value')
	})
})
