import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import request from 'supertest'

import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'

const namespace = basename(__filename)

describe('Test filesystem tree read operations', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateDirV2(it, expect, namespace, '/', 'dir1')
	helpers.itShouldCreateDirV2(it, expect, namespace, '/', 'dir2')

	it(`should read root dir with two paths`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/files-tree`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			data: {
				file: {
					path: '/',
					type: 'directory',
					createdAt: expect.stringMatching(regex.timestampRegex),
					updatedAt: expect.stringMatching(regex.timestampRegex),
				},
				paths: [
					{
						path: '/dir1',
						type: 'directory',
						createdAt: expect.stringMatching(regex.timestampRegex),
						updatedAt: expect.stringMatching(regex.timestampRegex),
					},
					{
						path: '/dir2',
						type: 'directory',
						createdAt: expect.stringMatching(regex.timestampRegex),
						updatedAt: expect.stringMatching(regex.timestampRegex),

					},
				],
			},
		})
	})

	helpers.itShouldCreateFileV2(it, expect, namespace,
		'/dir1',
		'foo1',
		'workflow',
		'text/plain',
		btoa(helpers.dummyWorkflow('foo1')))

	it(`should read root /dir1 with one path`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/files-tree/dir1`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			data: {
				file: {
					path: '/dir1',
					type: 'directory',
					createdAt: expect.stringMatching(regex.timestampRegex),
					updatedAt: expect.stringMatching(regex.timestampRegex),
				},
				paths: [
					{
						path: '/dir1/foo1',
						type: 'workflow',
						createdAt: expect.stringMatching(regex.timestampRegex),
						updatedAt: expect.stringMatching(regex.timestampRegex),
						size: 134,
					},
				],
			},
		})
	})
})
