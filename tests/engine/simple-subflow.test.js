import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'simplesubflowtest'

describe('Test subflow behaviour', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	helpers.itShouldCreateDir(it, expect, namespaceName, '/', 'a')

	helpers.itShouldCreateFile(it, expect, namespaceName,
		'/a',
		'child.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: noop
  transform:
    result: 'jq(.input + 1)'`))

	helpers.itShouldCreateFile(it, expect, namespaceName,
		'/a',
		'parent1.yaml',
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: child
  type: subflow
  workflow: '/a/child.yaml'
states:
- id: a
  type: action
  action:
    function: child
    input: 
      input: 1
  transform:
    result: 'jq(.return.result)'
`))

	it(`should invoke the '/a/parent1.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${namespaceName}/instances?path=a%2Fparent1.yaml&wait=true`)
		console.log(req.body)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			result: 2,
		})
	})

	helpers.itShouldCreateFile(it, expect, namespaceName,
		'/a',
		'parent2.yaml',
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: child
  type: subflow
  workflow: 'child.yaml'
states:
- id: a
  type: action
  action:
    function: child
    input: 
      input: 1
  transform:
    result: 'jq(.return.result)'
`))

	it(`should invoke the '/a/parent2.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${namespaceName}/instances?path=a%2Fparent2.yaml&wait=true`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			result: 2,
		})
	})

	helpers.itShouldCreateFile(it, expect, namespaceName,
		'/a',
		'parent3.yaml',
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: child
  type: subflow
  workflow: './child.yaml'
states:
- id: a
  type: action
  action:
    function: child
    input: 
      input: 1
  transform:
    result: 'jq(.return.result)'
`))

	it(`should invoke the '/a/parent3.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${namespaceName}/instances?path=a%2Fparent3.yaml&wait=true`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			result: 2,
		})
	})

	helpers.itShouldCreateFile(it, expect, namespaceName,
		'/a',
		'parent4.yaml',
		'workflow',
		'text/plain',
		btoa(`
functions:
- id: child
  type: subflow
  workflow: '../a/child.yaml'
states:
- id: a
  type: action
  action:
    function: child
    input: 
      input: 1
  transform:
    result: 'jq(.return.result)'
`))

	it(`should invoke the '/a/parent4.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${namespaceName}/instances?path=a%2Fparent4.yaml&wait=true`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			result: 2,
		})
	})

	it(`check if instances are present`, async () => {
		const instances = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${namespaceName}/instances`)
		expect(instances.statusCode).toEqual(200)
		expect(instances.body.meta.total).not.toBeLessThan(1)
	})

	// it(`check if instance logs are present`, async () => {
	// 	const instances = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances`)
	// 	// create a new array containing the ids
	// 	const ids = instances.body.instances.results.map(result => result.id)

	// 	// iterate over that array
	// 	await Promise.all(ids.map(async id => {
	// 		const logsResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances/${ id }/logs`)
	// 		expect(logsResponse.statusCode).toEqual(200)
	// 		expect(logsResponse.body.results.length).not.toBeLessThan(1)
	// 	}))
	// })

	// // TODO: Enable this test after new logging system merge.
	// it.skip(`check if namespace logs contains some workflow operations`, async () => {
	// 	const logsResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/logs`)
	// 	expect(logsResponse.statusCode).toEqual(200)
	// 	expect(logsResponse.body.results.length).not.toBeLessThan(1)
	// })
})
