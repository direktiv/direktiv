import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../../common'
import helpers from '../../common/helpers'
import request from '../../common/request'

const namespaceName = 'callpathtest'

describe('Test subflow behaviour', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	helpers.itShouldCreateDir(it, expect, namespaceName, '/', 'a')

	helpers.itShouldCreateFile(it, expect, namespaceName,
		'/a',
		`child.yaml`,
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: noop
  transform:
    result: 'jq(.input + 1)'`))

	helpers.itShouldCreateFile(it, expect, namespaceName,
		'/a',
		`parent1.yaml`,
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: child
  type: subflow
  workflow: '/a/child.yaml'
states:
- id: a
  type: action
  action:
    function: child
    input: 
      input: 1
  transform:
    result: 'jq(.return.result)'
`))

	it(`should invoke the '/a/parent1.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=a%2Fparent1.yaml&wait=true`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			result: 2,
		})
	})
})
