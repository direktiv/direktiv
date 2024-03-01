import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

const namespace = basename(__filename)

describe('Test filesystem tree read operations', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should read empty root dir`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/files`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			data: {
				path: '/',
				type: 'directory',
				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
				children: [],
			},
		})
	})

	helpers.itShouldCreateDirV2(it, expect, namespace, '','dir1')
	helpers.itShouldCreateDirV2(it, expect, namespace, '','dir2')
	helpers.itShouldCreateYamlFileV2(it, expect, namespace, '/', 'foo.yaml', 'workflow', helpers.dummyWorkflow('foo'))
	helpers.itShouldCreateYamlFileV2(it, expect, namespace, '/dir1', 'foo11.yaml', 'workflow', helpers.dummyWorkflow('foo11'))
	helpers.itShouldCreateYamlFileV2(it, expect, namespace, '/dir1', 'foo12.yaml', 'workflow', helpers.dummyWorkflow('foo12'))

	it(`should read root dir with three paths`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/files`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
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
					{
						path: '/foo.yaml',
						type: 'workflow',
						mimeType: 'application/direktiv',
						createdAt: expect.stringMatching(regex.timestampRegex),
						updatedAt: expect.stringMatching(regex.timestampRegex),

					},
				],
			},
		})
	})

	it(`should read dir1 with two files`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/files/dir1`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			data: {
				path: '/dir1',
				type: 'directory',
				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
				children: [
					{
						mimeType: 'application/direktiv',
						path: '/dir1/foo11.yaml',
						type: 'workflow',
						createdAt: expect.stringMatching(regex.timestampRegex),
						updatedAt: expect.stringMatching(regex.timestampRegex),

					},
					{
						mimeType: 'application/direktiv',
						path: '/dir1/foo12.yaml',
						type: 'workflow',
						createdAt: expect.stringMatching(regex.timestampRegex),
						updatedAt: expect.stringMatching(regex.timestampRegex),

					},
				],
			},
		})
	})

	it(`should read dir2 with zero files`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/files/dir2`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			data: {
				path: '/dir2',
				type: 'directory',
				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
				children: [],
			},
		})
	})

	helpers.itShouldDeleteFile(it, expect, namespace, '/foo.yaml')

	it(`should read root dir two dirs`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/files`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
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

	helpers.itShouldDeleteFile(it, expect, namespace, '/dir2')

	it(`should read root dir one path`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/files`)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
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
				],
			},
		})
	})

	it(`should read root not found`, async () => {
		const res = await request(config.getDirektivHost())
			.get(`/api/v2/namespaces/${ namespace }/files/dir2`)
		expect(res.statusCode).toEqual(404)
	})
})
