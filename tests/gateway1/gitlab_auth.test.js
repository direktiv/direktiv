import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10 } from '../common/retry'
import helpers from "../common/helpers.js";

const testNamespace = 'gitlab-auth'

describe('Test gitlab auth plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	helpers.itShouldTSWorkflow(
		it,
		expect,
		testNamespace,
		'/',
		'foo.wf.ts',
		`
function stateFirst(input) {
	return finish(input)
}
`,
	)


	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/',
		'endpoint.yaml',
		'endpoint',
		`x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/target"
    allow_anonymous: false
    plugins:
       auth:
       - type: gitlab-webhook-auth
         configuration:
            secret: secret
       target:
         type: target-flow
         configuration:
            flow: /foo.wf.ts
            content_type: application/json
post:
   responses:
      "200":
        description: works`,
	)

	retry10(`should execute`, async () => {
		const req = await request(common.config.getDirektivBaseUrl())
			.post(`/ns/` + testNamespace + `/target`)
			.set('X-Gitlab-Token', 'secret')
			.send({ hello: 'world' })

		expect(req.statusCode).toEqual(200)
	})

	retry10(`should fail`, async () => {
		const req = await request(common.config.getDirektivBaseUrl())
			.post(`/ns/` + testNamespace + `/target`)
			.set('X-Gitlab-Token', 'wrongsecret')
			.send({ hello: 'world' })

		expect(req.statusCode).toEqual(403)
	})
})
