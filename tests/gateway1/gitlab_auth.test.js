import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10 } from '../common/retry'

const testNamespace = 'gitlab-auth'

const wf = `
direktiv_api: workflow/v1
states:
- id: helloworld
  type: noop
  transform:
    result: jq(.)
`

const endpointFile = `x-direktiv-api: endpoint/v2
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
            flow: /target.yaml
            content_type: application/json
post:
   responses:
      "200":
        description: works`

describe('Test gitlab auth plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/',
		'target.yaml',
		'workflow',
		wf,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/',
		'endpoint.yaml',
		'endpoint',
		endpointFile,
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
