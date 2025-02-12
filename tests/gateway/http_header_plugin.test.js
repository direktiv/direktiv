import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10 } from '../common/retry'

const testNamespace = 'headers'

const endpointJSFile = `x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/target"
    allow_anonymous: true
    plugins:
       inbound:
       - type: header-manipulation
         configuration:
           headers_to_add:
           - name: hello
             value: world
           headers_to_modify: 
           - name: header1
             value: newvalue
           headers_to_remove:
             - name: header 
       - type: "request-convert"
         configuration:
           omit_headers: false
           omit_queries: true
           omit_body: true
           omit_consumer: true
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

describe('Test header plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

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
			.set('Header', 'Value1')
			.set('Header1', 'oldvalue')
			.send({ hello: 'world' })

		expect(req.statusCode).toEqual(200)
		expect(req.body.result.headers.Hello[0]).toEqual('world')
		expect(req.body.result.headers.Header).toBeUndefined()
		expect(req.body.result.headers.Header1[0]).toEqual('newvalue')
	})
})
