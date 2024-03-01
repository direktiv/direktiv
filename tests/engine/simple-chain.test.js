import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import helpers from "../common/helpers";

const namespaceName = 'simplechaintest'


describe('Test a simple chain of noop states', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	it(`should create a namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }`)

		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			namespace: {
				createdAt: expect.stringMatching(common.regex.timestampRegex),
				updatedAt: expect.stringMatching(common.regex.timestampRegex),
				name: namespaceName,
			},
		})
	})

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'simple-chain.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: noop
  transform:
    a: x
  transition: b
- id: b
  type: noop
  transform: 'jq(.b = "y")'
  transition: c
- id: c
  type: noop
  transform: 'jq(.c = "z")'
`))

	it(`should invoke the '/simple-chain.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/simple-chain.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			a: 'x',
			b: 'y',
			c: 'z',
		})
	})

})
