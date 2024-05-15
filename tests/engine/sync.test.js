import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'synctest'

describe('Test synchronous behaviour', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'noop.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: noop
`))

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'parallel.yaml',
		'workflow',
		'text/plain',
		btoa(`
direktiv_api: workflow/v1
description: A simple 'no-op' state that returns 'Hello world!'
functions:
- type: subflow
  id: sub
  workflow: /noop.yaml
states:
- id: a
  type: parallel
  mode: and
  timeout: PT2M
  actions:
  - function: sub
    input:
      x: 0
  - function: sub
    input:
      x: 1
  - function: sub
    input:
      x: 2
  - function: sub
    input:
      x: 3
  - function: sub
    input:
      x: 4
  - function: sub
    input:
      x: 5
  - function: sub
    input:
      x: 6
  - function: sub
    input:
      x: 7
  - function: sub
    input:
      x: 8
  - function: sub
    input:
      x: 9
  - function: sub
    input:
      x: 10
  - function: sub
    input:
      x: 11
`))

	it(`should invoke the '/parallel.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespaceName }/instances?path=parallel.yaml&wait=true`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			return: [
				{ x: 0 },
				{ x: 1 },
				{ x: 2 },
				{ x: 3 },
				{ x: 4 },
				{ x: 5 },
				{ x: 6 },
				{ x: 7 },
				{ x: 8 },
				{ x: 9 },
				{ x: 10 },
				{ x: 11 },
			],
		})
	})
})
