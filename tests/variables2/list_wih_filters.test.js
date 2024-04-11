import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'

const namespace = basename(__filename)

describe('Test variable list calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateVariableV2(it, expect, namespace, {
		name: 'foo1',
		data: btoa('foo1'),
		mimeType: 'mime_foo1',
	})

	helpers.itShouldCreateVariableV2(it, expect, namespace, {
		name: 'foo2',
		data: btoa('foo2'),
		mimeType: 'mime_foo2',
	})

	it(`should list variable all`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables`)
		expect(res.statusCode).toEqual(200)

		const reduced = res.body.data.map(item => item.name)
		expect(reduced).toEqual([ 'foo1', 'foo2' ])
	})

	it(`should list variable foo2`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables?name=foo2`)
		expect(res.statusCode).toEqual(200)

		const reduced = res.body.data.map(item => item.name)
		expect(reduced).toEqual([ 'foo2' ])
	})

	it(`should list empty`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/variables?name=foo3`)
		expect(res.statusCode).toEqual(200)

		const reduced = res.body.data.map(item => item.name)
		expect(reduced).toEqual([])
	})
})
