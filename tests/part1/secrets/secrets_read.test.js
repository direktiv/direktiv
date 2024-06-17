import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../../common'
import helpers from '../../common/helpers'
import request from '../../common/request'

const testNamespace = 'test-secrets-namespace'
const testWorkflow = 'test-secret'

describe('Test secret read operations', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, testNamespace)

	it(`should create a new secret`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${ testNamespace }/secrets`)
			.send({
				name: 'key1',
				data: btoa('value1'),
			})
		expect(res.statusCode).toEqual(200)
	})

	helpers.itShouldCreateFile(it, expect, testNamespace,
		'',
		`${ testWorkflow }-parent.yaml`,
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: echo
  workflow: ${ testWorkflow }-child.yaml
  type: subflow
states:
- id: echo
  type: action
  action:
    function: echo
    secrets: [key1]
    input: 
      secret: 'jq(.secrets.key1)'
  transform: 
    result: 'jq(.return.secret)'
`))

	helpers.itShouldCreateFile(it, expect, testNamespace,
		'',
		`${ testWorkflow }-child.yaml`,
		'workflow',
		'text/plain',
		btoa(`
states:
- id: helloworld
  type: noop
`))

	it(`should invoke the '/${ testWorkflow }-parent.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ testNamespace }/instances?path=${ testWorkflow }-parent.yaml&wait=true`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			result: 'value1',
		})
	})
})
