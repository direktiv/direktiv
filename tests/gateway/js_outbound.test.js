import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10 } from '../common/retry'

const testNamespace = 'js-outbound'

const endpointJSFile = `x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/target"
    allow_anonymous: true
    plugins:
      outbound:
      - type: js-outbound
        configuration:
          script: |
            input["Headers"].Add("Header2", "value2")
            b = JSON.parse(input["Body"])
            b["random"] = "data"
            input["Body"] = JSON.stringify(b)
            input["Code"] = 201
      target:
        type: target-flow
        configuration:
          flow: /target.yaml
          content_type: application/json
post:
   responses:
      "200":
        description: works`

const wf = `
direktiv_api: workflow/v1
states:
- id: helloworld
  type: noop
  transform:
    result: jq(.)
`

describe('Test js outbound plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)
	// common.helpers.itShouldCreateNamespace(it, expect, testNamespace);

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		endpointJSFile,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'target.yaml', 'workflow',
		wf,
	)

	retry10(`should have expected body after js`, async () => {
		const req = await request(common.config.getDirektivHost()).post(
			`/ns/` + testNamespace + `/target?Query1=value1&Query2=value2`,
		)
			.set('Header1', 'Value1')
			.send({ hello: 'world' })
		expect(req.statusCode).toEqual(201)

		// added header in the script
		expect(req.header.header2).toEqual('value2')

		// added random data in the script
		expect(req.body.random).toEqual('data')
	})
})
