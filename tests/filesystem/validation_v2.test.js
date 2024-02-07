import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import request from 'supertest'

import config from '../common/config'
import helpers from '../common/helpers'

const namespace = basename(__filename)

describe('Test filesystem tree read operations', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should fail creating file with invalid base64 data`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/files-tree`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'foo',
				type: 'workflow',
				mimeType: 'text/plain',
				data: 'some_invalid_base64',
			})
		expect(res.statusCode).toEqual(400)
		expect(res.body).toMatchObject({
			error: {
				code: 'request_data_invalid',
				message: 'file data has invalid base64 string',
			},
		})
	})

	it(`should fail creating file with invalid yaml data`, async () => {
		const res = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/files-tree`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'foo',
				type: 'workflow',
				mimeType: 'text/plain',
				data: btoa('11 some_invalid_yaml 11 \nsome_invalid_yaml'),
			})
		expect(res.statusCode).toEqual(400)
		expect(res.body).toMatchObject({
			error: {
				code: 'request_data_invalid',
				message: 'file data has invalid yaml string',
			},
		})
	})
})
