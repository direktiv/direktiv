import { beforeAll, describe, expect, it } from '@jest/globals'
import { btoa } from 'js-base64'
import { basename } from 'path'

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const namespace = basename(__filename)

describe('Test target-flow plugin', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	it(`should set plain text variable`, async () => {
		const workflowVarResponse = await request(common.config.getDirektivBaseUrl()).post(`/api/v2/namespaces/${ namespace }/variables`)
			.send({
				name: 'foo',
				data: btoa('Hello World'),
				mimeType: 'text/plain',
			})
		expect(workflowVarResponse.statusCode).toEqual(200)
	})

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep1.yaml', 'endpoint', `
    x-direktiv-api: endpoint/v2
    x-direktiv-config:
        path: "/ep1"
        allow_anonymous: true
        plugins:
          target:
            type: target-namespace-var
            configuration:
                namespace: ${ namespace }
                variable: foo
    get:
      responses:
         "200":
           description: works
`)

	retry10(`should execute wf1.yaml file`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(`/api/v2/namespaces/${ namespace }/gateway/ep1`)
		expect(res.statusCode).toEqual(200)
		expect(res.text).toEqual('Hello World')
		expect(res.headers['content-type']).toEqual('text/plain')
		expect(res.headers['content-length']).toEqual('11')
	})
})
