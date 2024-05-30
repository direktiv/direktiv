import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'simplechaintest'

describe('Test a simple chain of noop states', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	helpers.itShouldCreateFile(it, expect, namespaceName,
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
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=simple-chain.yaml&wait=true`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			a: 'x',
			b: 'y',
			c: 'z',
		})
	})
})
