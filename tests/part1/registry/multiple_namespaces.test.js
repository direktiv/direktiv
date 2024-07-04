import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../../common'
import request from '../../common/request'

describe('Test services crud operations', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, 'test_namespace_a')
	common.helpers.itShouldCreateNamespace(it, expect, 'test_namespace_b')

	itShouldCreateSecret(it, expect, 'test_namespace_a', 'a_domain_1.io', 'a_name_1', 'a_password1')
	itShouldCreateSecret(it, expect, 'test_namespace_a', 'a_domain_2.io', 'a_name_2', 'a_password2')
	itShouldCreateSecret(it, expect, 'test_namespace_b', 'b_domain_1.io', 'b_name_1', 'b_password1')
	itShouldCreateSecret(it, expect, 'test_namespace_b', 'b_domain_2.io', 'b_name_2', 'b_password2')

	it(`should list all registries in namespace test_namespace_a`, async () => {
		const listRes = await request(common.config.getDirektivHost())
			.get(`/api/v2/namespaces/test_namespace_a/registries`)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data).toEqual(
			expect.arrayContaining(

				[
					{
						namespace: 'test_namespace_a',
						id: expect.stringMatching(/^secret-/),
						url: 'a_domain_1.io',
						user: 'a_name_1',
					},
					{
						namespace: 'test_namespace_a',
						id: expect.stringMatching(/^secret-/),
						url: 'a_domain_2.io',
						user: 'a_name_2',
					}
				]

			)
		)
	})

	it(`should list all registries in namespace test_namespace_b`, async () => {
		const listRes = await request(common.config.getDirektivHost())
			.get(`/api/v2/namespaces/test_namespace_b/registries`)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data).toEqual(
			expect.arrayContaining([
				{
					namespace: 'test_namespace_b',
					id: expect.stringMatching(/^secret-/),
					url: 'b_domain_1.io',
					user: 'b_name_1',
				},
				{
					namespace: 'test_namespace_b',
					id: expect.stringMatching(/^secret-/),
					url: 'b_domain_2.io',
					user: 'b_name_2',
				}],
			)
		)
	})
})

function itShouldCreateSecret(it, expect, namespace, url, user, password) {
	it(`should create a registry ${url} ${user} ${password}`, async () => {
		const res = await request(common.config.getDirektivHost())
			.post(`/api/v2/namespaces/${namespace}/registries`)
			.send({
				url,
				user,
				password,
			})
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: {
				namespace,
				id: expect.stringMatching(/^secret-/),
				url,
				user,
			},
		})
	})
}
