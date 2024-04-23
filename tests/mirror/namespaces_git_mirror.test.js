import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import {basename} from "path";
import {retry50} from "../common/retry";

const namespace = basename(__filename)

describe('Test namespace git mirroring', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	it(`should create a new git mirrored namespace`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces`)
			.send({
				name: namespace,
				mirror: {
					url: 'https://github.com/direktiv/direktiv-examples.git',
					gitRef: 'main',
					authType: "public",
				}
			})
		expect(res.statusCode).toEqual(200)
	})
	it(`should trigger a new sync`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${namespace}/syncs`)
			.send({})
		expect(res.statusCode).toEqual(200)
	})
	retry50(`should get the new git namespace`, async () => {
		const res = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/files/aws`)
		expect(res.statusCode).toEqual(200)
	})
	it(`should delete the new git namespace`, async () => {
		const res = await request(common.config.getDirektivHost()).delete(`/api/v2/namespaces/${ namespace }`)
		expect(res.statusCode).toEqual(200)
	})
	it(`should get 404 after the new git namespace deletion`, async () => {
		const res = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/files/aws`)
		expect(res.statusCode).toEqual(404)
	})
})
