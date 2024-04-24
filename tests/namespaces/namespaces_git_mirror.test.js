import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../common'
import regex from '../common/regex'
import request from '../common/request'
import { retry50 } from '../common/retry'

const namespace = basename(__filename)

describe('Test namespace git mirroring', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	it(`should create a new git mirrored namespace`, async () => {
		const res = await request(common.config.getDirektivHost())
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
		const res = await request(common.config.getDirektivHost())
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
		const res = await request(common.config.getDirektivHost())
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
		const res = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/files/listener.yml`)
		expect(res.statusCode).toEqual(200)
	})

	it(`should read the workflow variables of '/banana/page-1.yaml'`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/variables?workflowPath=/banana/page-1.yaml`)
		expect(req.statusCode).toEqual(200)
		expect(req.body.data).toEqual([ {
			id: expect.stringMatching(common.regex.uuidRegex),
			type: 'workflow-variable',
			reference: '/banana/page-1.yaml',
			name: 'page.html',
			size: 221,
			mimeType: 'text/html',
			createdAt: expect.stringMatching(regex.timestampRegex),
			updatedAt: expect.stringMatching(regex.timestampRegex),
		} ])
	})

	it(`should read the workflow variables of '/banana/page-2.yaml'`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/variables?workflowPath=/banana/page-2.yaml`)
		expect(req.statusCode).toEqual(200)
		expect(req.body.data).toEqual([ {
			id: expect.stringMatching(common.regex.uuidRegex),
			type: 'workflow-variable',
			reference: '/banana/page-2.yaml',
			name: 'Page.HTML',
			size: 233,
			mimeType: 'text/html',
			createdAt: expect.stringMatching(regex.timestampRegex),
			updatedAt: expect.stringMatching(regex.timestampRegex),
		} ])
	})

	it(`should check for the expected list of namespace variables`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/variables`)
		expect(req.statusCode).toEqual(200)
		const reduced = req.body.data.map(i => i.name)
		expect(reduced.sort()).toEqual([
			'beta.json',
			'ALPHA.json',
			'alp_ha.json',
			'data.json',
			'alpha.json',
			'alp-ha.json',
			'alpha.csv',
			'alpha_.json',
			'gamma.css',
		].sort())
	})

	it(`should delete the new git namespace`, async () => {
		const res = await request(common.config.getDirektivHost()).delete(`/api/v2/namespaces/${ namespace }`)
		expect(res.statusCode).toEqual(200)
	})

	it(`should get 404 after the new git namespace deletion`, async () => {
		const res = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/files/listener.yml`)
		expect(res.statusCode).toEqual(404)
	})
})
