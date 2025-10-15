import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'
import { retry10 } from '../common/retry'

const testNamespace = 'headers'

describe('Test header plugin', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'endpoint1.yaml', 'endpoint',
		`
x-direktiv-api: endpoint/v2
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
            flow: /foo.wf.ts
            content_type: application/json
post:
   responses:
      "200":
        description: works`,
	)

	common.helpers.itShouldCreateYamlFile(
		it,
		expect,
		testNamespace,
		'/', 'foo.wf.ts', 'workflow',
		`
function stateFirst(input) {
	return finish(input)
}
`,
	)

	retry10(`should have expected body after js`, async () => {
		const req = await request(common.config.getDirektivBaseUrl()).post(
			`/ns/` + testNamespace + `/target?Query1=value1&Query2=value2`,
		)
			.set('Header', 'Value1')
			.set('Header1', 'oldvalue')
			.send({ hello: 'world' })

		const got = JSON.parse(req.body.data.output)

		expect(req.statusCode).toEqual(200)
		expect(got.headers.Hello[0]).toEqual('world')
		expect(got.headers.Header).toBeUndefined()
		expect(got.headers.Header1[0]).toEqual('newvalue')
		expect(got.method).toEqual('POST')
	})
})
