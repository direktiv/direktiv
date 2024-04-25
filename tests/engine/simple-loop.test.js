import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'simplelooptest'

describe('Test a simple loop', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'simple-loop.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: init
  type: noop
  transform:
    i: 5
    result: []
  transition: check
- id: check
  type: switch
  conditions:
  - condition: 'jq(.i)'
    transform: 'jq(.result += [.i] | .i -= 1)'
    transition: check
`))

	it(`should invoke the '/simple-loop.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/simple-loop.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			i: 0,
			result: [ 5, 4, 3, 2, 1 ],
		})
	})
})
