import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import { basename } from 'path'

const namespace = basename(__filename)
const testWorkflow = 'test-workflow.yaml'

describe('Test workflow echo json action', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateFile(it, expect, namespace,
		'',
		testWorkflow,
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: echo
  image: gcr.io/direktiv/functions/echo:1.0
  type: knative-workflow
states:
- id: echo
  type: action
  action:
    function: echo
`))

	it(`should invoke the ${testWorkflow} workflow and echo input`, async () => {
		await helpers.sleep(500)
		const res = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${namespace}/instances?path=${testWorkflow}&wait=true`).send('{"hello":"world"}')
		expect(res.statusCode).toEqual(200)
		expect(res.body.return).toMatchObject({
			"hello": "world"
		})
	})
	it(`should invoke the ${testWorkflow} workflow  and echo input run 2`, async () => {
		const res = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${namespace}/instances?path=${testWorkflow}&wait=true`).send('{"hello2":"world"}')
		expect(res.statusCode).toEqual(200)
		expect(res.body.return).toMatchObject({
			"hello2": "world"
		})
	})

})
