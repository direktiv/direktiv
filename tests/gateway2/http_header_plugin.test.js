import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const namespace = basename(__filename)

describe('Test header plugin', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
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

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep1.yaml', 'endpoint', `
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
path: /target
`)

	helpers.itShouldCreateYamlFile(
		it,
		expect,
		namespace,
		'/', 'target.yaml', 'workflow', `
direktiv_api: workflow/v1
states:
- id: helloworld
  type: noop
  transform:
    result: jq(.)`)

	retry10(`should have expected body after js`, async () => {
		const req = await request(common.config.getDirektivHost()).post(
			`/api/v2/namespaces/${ namespace }/gateway2/target?Query1=value1&Query2=value2`,
		)
			.set('Header', 'Value1')
			.set('Header1', 'oldvalue')
			.send({ hello: 'world' })
			.auth('user1', 'pwd1')

		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
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
