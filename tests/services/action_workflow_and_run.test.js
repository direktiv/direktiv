import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const testNamespace = 'test-services'
const testWorkflow = 'test-workflow.yaml'

describe('Test workflow function invoke', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, testNamespace)

	helpers.itShouldCreateFileV2(it, expect, testNamespace,
		'',
		testWorkflow,
		'workflow',
		'text/plain',
		btoa(`
description: A simple 'action' state that sends a get request
functions:
- id: get
  image: direktiv/request:v4
  type: knative-workflow
states:
- id: getter 
  type: action
  action:
    function: get
    input: 
      method: "GET"
      url: "https://jsonplaceholder.typicode.com/todos/1"
`))

	it(`should invoke the ${ testWorkflow } workflow`, async () => {
		await helpers.sleep(500)
		const res = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ testNamespace }/tree/${ testWorkflow }?op=wait`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.return.status).toBe('200 OK')
	})
})
