import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'
import common from "../common";

const namespace = basename(__filename)

describe('Test gateway duplicated endpoint path', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep1.yaml', 'endpoint', `
direktiv_api: endpoint/v1
path: /foo
allow_anonymous: true
methods:
  - POST
plugins:
  target:
    type: debug-target
`)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep2.yaml', 'endpoint', `
direktiv_api: endpoint/v1
path: /foo
allow_anonymous: true
methods:
  - POST
plugins:
  target:
    type: debug-target
`)

	retry10(`should execute gateway ep1.yaml endpoint`, async () => {
		const res = await request(config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/gateway/foo`)
			.send({})
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.text).toEqual('from debug plugin')
	})

	retry10(`should have error set in the second endpoint`, async () => {
		const listRes = await request(common.config.getDirektivHost()).get(
			`/api/v2/namespaces/${ namespace }/gateway/routes`,
		)
		expect(listRes.statusCode).toEqual(200)
		expect(listRes.body.data.length).toEqual(2)
		listRes.body.data[0].errors = []
		listRes.body.data[1].errors = [ 'duplicate gateway path: /foo']
	})
})
