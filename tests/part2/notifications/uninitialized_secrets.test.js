import { beforeAll, describe, expect, it } from '@jest/globals'

import { basename } from 'path'
import config from '../../common/config'
import { fileURLToPath } from 'url'
import helpers from '../../common/helpers'
import request from '../../common/request'

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test uninitialized secrets notifications', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should read no notifications`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${namespace}/notifications`,
		)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [],
		})
	})

	helpers.itShouldCreateFile(
		it,
		expect,
		namespace,
		'/',
		'secrets.wf.ts',
		'workflow',
		'application/typescript',
		btoa(`
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT5S",
  state: "stateCreateSecrets",
};

function stateCreateSecrets(): StateFunction<unknown> {
  // this initializes the secrets if they don't exist yet
  getSecrets(["foo", "bar"]);

  return finish("uninitialized secrets should now exist");
}
`),
	)

	it(`should read one notification`, async () => {
		await helpers.sleep(500)

		const res = await request(config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${namespace}/notifications`,
		)
		expect(res.statusCode).toEqual(200)
		expect(res.body).toEqual({
			data: [
				{
					description: 'secrets have not been initialized: [bar foo]',
					count: 2,
					level: 'warning',
					type: 'uninitialized_secrets',
				},
			],
		})
	})
})
