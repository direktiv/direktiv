import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../../common'
import helpers from '../../common/helpers'
import request from '../../common/request'

const testNamespace = 'test-secrets-namespace'
const testWorkflow = 'test-secret'

describe('Test secret read operations', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, testNamespace)

	it(`should create a new secret`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.post(`/api/v2/namespaces/${testNamespace}/secrets`)
			.send({
				name: 'key-one',
				data: btoa('value1'),
			})
		expect(res.statusCode).toEqual(200)
	})

	helpers.itShouldCreateFile(
		it,
		expect,
		testNamespace,
		'',
		`${testWorkflow}.wf.ts`,
		'workflow',
		'application/typescript',
		btoa(`
const flow: FlowDefinition = {
  type: "default",
  timeout: "PT5S",
  state: "stateEchoSecret",
};

function stateEchoSecret() {
  const secret = getSecret("key-one");

  return finish(secret);
}
`),
	)

	it(`should invoke the '/${testWorkflow}.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivBaseUrl()).post(
			`/api/v2/namespaces/${testNamespace}/instances?path=${testWorkflow}.wf.ts&wait=true`,
		)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			data: 'value1',
		})
	})
})
