import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'
import common from "../common";
import {btoa} from "js-base64";

const namespace = basename(__filename)

describe('Test target-flow plugin', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should set plain text variable`, async () => {
		const workflowVarResponse = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/variables`)
			.send({
				name: 'foo',
				data: btoa('Hello World'),
				mimeType: 'text/plain',
			})
		expect(workflowVarResponse.statusCode).toEqual(200)
	})

	helpers.itShouldCreateYamlFileV2(it, expect, namespace,
		'/', 'ep1.yaml', 'endpoint', `
direktiv_api: endpoint/v2
path: /ep1
methods: 
  - GET
allow_anonymous: true
plugins:
  target:
    type: target-namespace-var
    configuration:
        namespace: ${ namespace }
        variable: foo
`)
	retry10(`should execute wf1.yaml file`, async () => {
		const res = await request(config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/gateway2/ep1`)
		expect(res.statusCode).toEqual(200)

		expect(res.body.data).toMatchObject({
			data: btoa("Hello World"),
		})
	})
})
