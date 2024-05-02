import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../../common/config'
import helpers from '../../common/helpers'
import request from '../../common/request'

const namespace = basename(__filename)

describe('Test uninitialized secrets notifications', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should read no notifications`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/notifications`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [],
		})
	})

	helpers.itShouldCreateFileV2(it, expect, namespace,
		'/',
		'foo1',
		'workflow',
		'text/plain',
		btoa(`
direktiv_api: workflow/v1
functions:
- type: subflow
  id: myfunc
  workflow: subflow.yaml
states:
- id: a
  type: action
  action:
    function: myfunc
    input: 'jq(.x)'
    secrets: ["a", "b"]
`))

	it(`should read one notification`, async () => {
		await helpers.sleep(500)

		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/notifications`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [ {
				description: 'secrets have not been initialized: [a b]',
				count: 2,
				level: 'warning',
				type: 'uninitialized_secrets',
			} ],
		})
	})
})
