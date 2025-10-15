import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import { fileURLToPath } from 'url'

import common from '../../common'
import regex from '../../common/regex'
import request from '../../common/request'
import { retry50 } from '../../common/retry'

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test namespace git mirroring', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	it(`should create a new git mirrored namespace`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces`)
			.send({
				name: namespace,
				mirror: {
					url: 'https://github.com/direktiv/direktiv-test-project.git',
					gitRef: 'main',
					authType: 'public',
				},
			})
		expect(res.statusCode).toEqual(200)
	})

	it(`should trigger a new sync`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${ namespace }/syncs`)
			.send({})
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: {
				id: expect.stringMatching(common.regex.uuidRegex),
				status: 'pending',
				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
				endedAt: expect.stringMatching(regex.timestampRegex),
			},
		})
	})

	retry50(`should succeed to sync`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.get(`/api/v2/namespaces/${ namespace }/syncs`)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data).toEqual([
			{
				id: expect.stringMatching(common.regex.uuidRegex),
				status: 'complete',
				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
				endedAt: expect.stringMatching(regex.timestampRegex),
			},
		])
	})

	it(`should get the new git namespace`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).get(`/api/v2/namespaces/${ namespace }/files/listener.yml`)
		expect(res.statusCode).toEqual(200)
	})

	it(`should delete the new git namespace`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).delete(`/api/v2/namespaces/${ namespace }`)
		expect(res.statusCode).toEqual(200)
	})

	it(`should get 404 after the new git namespace deletion`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).get(`/api/v2/namespaces/${ namespace }/files/listener.yml`)
		expect(res.statusCode).toEqual(404)
	})
})
