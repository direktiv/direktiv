import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'
import {fileURLToPath} from "url";

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test header plugin', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'c1.yaml', 'consumer', `
direktiv_api: "consumer/v1"
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
x-direktiv-api: endpoint/v2
x-direktiv-config:
    path: "/target"
    allow_anonymous: false
    plugins:
        auth:
        - type: basic-auth  
        target:
            type: target-flow
            configuration:
                flow: /target.wf.ts
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
post:
    responses:
        "200":
        description: works
`)

	helpers.itShouldTSWorkflow(
		it,
		expect,
		namespace,
		'/', 'target.wf.ts', `
function stateFirst(input) {
	return finish(input)
}
`)

	retry10(`should have expected body after js`, async () => {
		const res = await request(common.config.getDirektivBaseUrl()).post(
			`/api/v2/namespaces/${ namespace }/gateway/target?Query1=value1&Query2=value2`,
		)
			.set('Header', 'Value1')
			.set('Header1', 'oldvalue')
			.send({ hello: 'world' })
			.auth('user1', 'pwd1')

		expect(res.statusCode).toEqual(200)
		const got = JSON.parse(res.body.data.output)
		const want = {
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
		}
		expect(got).toMatchObject(want)
	})
})
