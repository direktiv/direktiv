// TODO: tests for no, one and multiple gateway files
// TODO: tests for broken gateway file

import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'
import {fileURLToPath} from "url";

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test gateway no basic file', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	retry10(`should  get virtual config file`, async () => {
		const res = await request(config.getDirektivBaseUrl()).post(`/api/v2/namespaces/${ namespace }/gateway/info`)
			.send({})
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.file_path).toEqual('virtual')
		expect(res.body.data.spec.paths).toEqual({})
	})
})

describe('Test gateway basic file', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'gw.yaml', 'gateway', `
openapi: 3.0.0
x-direktiv-api: gateway/v1

info:
   title: mytitle
   version: myversion
`)

	retry10(`should  get virtual config file`, async () => {
		const res = await request(config.getDirektivBaseUrl()).post(`/api/v2/namespaces/${ namespace }/gateway/info`)
			.send({})
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.file_path).toEqual('/gw.yaml')
		expect(res.body.data.spec.info.title).toEqual('mytitle')
		expect(res.body.data.spec.info.version).toEqual('myversion')
	})
})

describe('Test gateway broken basic file', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'gw.yaml', 'gateway', `
openapi: 3.0.0
x-direktiv-api: gateway/v1

info:
   title: mytitle
   version: myversion

additional:
   key: value
`)

	// retry10(`should  get virtual config file`, async () => {
	// 	const res = await request(config.getDirektivBaseUrl()).post(`/api/v2/namespaces/${ namespace }/gateway/info`)
	// 		.send({})
	// 	expect(res.statusCode).toEqual(200)
	// 	expect(res.body.data.errors.length).toEqual(1)
	// })
})

describe('Test gateway with multiple basic files', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'gw.yaml', 'gateway', `
openapi: 3.0.0
x-direktiv-api: gateway/v1

info:
   title: mytitle
   version: myversion
`)

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'gw2.yaml', 'gateway', `
openapi: 3.0.0
x-direktiv-api: gateway/v1

info:
   title: mytitle
   version: myversion
`)

	// retry10(`should  get virtual config file`, async () => {
	// 	const res = await request(config.getDirektivBaseUrl()).post(`/api/v2/namespaces/${ namespace }/gateway/info`)
	// 		.send({})
	// 	expect(res.statusCode).toEqual(200)
	// 	expect(res.body.data.errors.length).toEqual(1)
	// 	expect(res.body.data.errors[0].startsWith('multiple gateway specifications found')).toBeTruthy()
	// })
})

// TODO: test broken spec
