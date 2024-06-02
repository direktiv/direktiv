import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const testNamespace = 'headers'

const endpointJSFile = `
direktiv_api: endpoint/v2
allow_anonymous: false
plugins:
  auth:
    - type: basic-auth   
  target:
    type: target-flow
    configuration:
        flow: /target.yaml
        content_type: application/json
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
methods: 
  - POST
path: /target`

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
	helpers.itShouldCreateYamlFile(it, expect, testNamespace,
		'/', 'c1.yaml', 'consumer', `
direktiv_api: "consumer/v2"
username: user1
password: pwd1
api_key: key1
tags:
- tag1
groups:
- group1
`)

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
			`/api/v2/namespaces/${ testNamespace }/gateway2/target?Query1=value1&Query2=value2`,
		)
			.set('Header', 'Value1')
			.set('Header1', 'oldvalue')
			.send({ hello: 'world' })
			.auth('user1', 'pwd1')

		expect(req.statusCode).toEqual(200)
		expect(req.body).toEqual({
			result: {
				body: {
					hello: 'world',
				},
				consumer: {
					groups: [
						'group1',
					],
					tags: [
						'tag1',
					],
					username: 'user1',
				},
				headers: {
					'Accept-Encoding': [
						'gzip, deflate',
					],
					Authorization: [
						'Basic dXNlcjE6cHdkMQ==',
					],
					Connection: [
						'close',
					],
					'Content-Length': [
						'17',
					],
					'Content-Type': [
						'application/json',
					],
					Header1: [
						'newvalue',
					],
					Hello: [
						'world',
					],
				},
				query_params: {
					Query1: [
						'value1',
					],
					Query2: [
						'value2',
					],
				},
				url_params: {},
			},
		},
		)
	})
})
