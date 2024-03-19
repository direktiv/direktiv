import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const testNamespace = 'test-secrets-namespace'
const testWorkflow = 'test-secret'

describe('Test secret read operations', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	it(`should create a new namespace`, async () => {
		const res = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ testNamespace }`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: {
				name: testNamespace,
				// regex /^2.*Z$/ matches timestamps like 2023-03-01T14:19:52.383871512Z
				createdAt: expect.stringMatching(/^2.*Z$/),
				updatedAt: expect.stringMatching(/^2.*Z$/),
			},
		})
	})

	it(`should create a new secret`, async () => {
		const res = await request(common.config.getDirektivHost())
			.put(`/api/namespaces/${ testNamespace }/secrets/key1`)
			.set({
				'Content-Type': 'text/plain',
			})

			.send(`value1`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: testNamespace,
			key: 'key1',
		})
	})

	helpers.itShouldCreateFileV2(it, expect, testNamespace,
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

	helpers.itShouldCreateFileV2(it, expect, testNamespace,
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
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ testNamespace }/tree/${ testWorkflow }-parent.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			result: 'value1',
		})
	})
})
