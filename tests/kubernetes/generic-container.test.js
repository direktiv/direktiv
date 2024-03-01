import { beforeAll, describe, expect, it } from '@jest/globals'

import request from '../common/request'
import { retry10 } from '../common/retry'
import config from "../common/config";
import helpers from "../common/helpers";


const testNamespace = 'patches'


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

	helpers.itShouldCreateFile(
		it,
		expect,
		testNamespace,
		'/wf1.yaml',
		genericContainerWorkflow,
	)


	retry10(`should invoke workflow`, async () => {
		const res = await request(config.getDirektivHost()).get(`/api/namespaces/${ testNamespace }/tree/wf1.yaml?op=wait`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.return[0].Output).toEqual('data')
	})


})

