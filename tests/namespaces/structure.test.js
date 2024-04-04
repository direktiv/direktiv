import { beforeAll, describe, expect, it } from '@jest/globals'

import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

const timestamps = {
	createdAt: expect.stringMatching(regex.timestampRegex),
	updatedAt: expect.stringMatching(regex.timestampRegex),
}

describe('Test namespace get delete list calls', () => {
	beforeAll(helpers.deleteAllNamespaces)

	it(`should create a new namespace foo`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces`)
			.send({
				name: 'foo',
			})
		expect(res.statusCode).toEqual(200)
	})

	it(`should get the new namespace foo`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/foo`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toEqual({
			name: 'foo',
			...timestamps,
		})
	})

	it(`should list the new namespace foo`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.length).toEqual(1)
		expect(res.body.data[0]).toEqual({
			name: 'foo',
			...timestamps,
		})
	})
})
