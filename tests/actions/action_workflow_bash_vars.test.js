import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import { basename } from 'path'

const namespace = basename(__filename)
const testWorkflow = 'test-workflow-bash.yaml'

describe('Test workflow bash vars via action', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateFile(it, expect, namespace,
		'',
		testWorkflow,
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: bash
  image: gcr.io/direktiv/functions/python:1.0
  type: knative-workflow
states:
- id: bash 
  type: action
  action:
    function: bash
    files:
    - key: myvar
      scope: workflow
`))
	it(`should set plain text variable`, async () => {
		const workflowVarResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${namespace}/variables`)
			.send({
				name: 'myvar',
				data: btoa('Hello World 55'),
				mimeType: 'text/plain',
				workflowPath: "/" + testWorkflow,
			})
		expect(workflowVarResponse.statusCode).toEqual(200)
	})
	it(`read variable via bash action from ${testWorkflow} workflow`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${namespace}/instances?path=${testWorkflow}&wait=true`)
			.send({
				commands: [
					{
						command: "cat myvar"
					}
				]
			});
		expect(res.statusCode).toEqual(200)

		expect(res.body.return.python).toMatchObject(
			[{ "result": "Hello World 55", "success": true }])
	})
	it(`set new variable via bash action from ${testWorkflow} workflow`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${namespace}/instances?path=${testWorkflow}&wait=true`)
			.send({
				commands: [
					{
						command: "bash -c 'echo hi > out/workflow/somevar1'",
					}
				]
			});
		expect(res.statusCode).toEqual(200)
		const workflowVarResponse = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${namespace}/variables?workflowPath=/${testWorkflow}`)

		expect(workflowVarResponse.statusCode).toEqual(200)
		expect(workflowVarResponse.body.data.length).toBe(2)

	})

})
