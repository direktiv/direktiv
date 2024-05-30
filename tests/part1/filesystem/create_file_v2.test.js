import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../../common/config'
import helpers from '../../common/helpers'
import regex from '../../common/regex'
import request from '../../common/request'

const namespace = basename(__filename)

describe('Test filesystem tree read operations', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateDir(it, expect, namespace, '/', 'dir1')
	helpers.itShouldCreateDir(it, expect, namespace, '/', 'dir2')

	it(`should read root dir with two paths`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/files`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: {
				path: '/',
				type: 'directory',
				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
				children: [
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

	helpers.itShouldCreateFile(it, expect, namespace,
		'/dir1',
		'foo1',
		'workflow',
		'text/plain',
		btoa(helpers.dummyWorkflow('foo1')))

	it(`should read root /dir1 with one path`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/files/dir1`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: {
				path: '/dir1',
				type: 'directory',
				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
				children: [
					{
						path: '/dir1/foo1',
						type: 'workflow',
						mimeType: 'text/plain',
						createdAt: expect.stringMatching(regex.timestampRegex),
						updatedAt: expect.stringMatching(regex.timestampRegex),
						size: 134,
					},
				],
			},
		})
	})
})
