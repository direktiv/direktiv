import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import request from '../common/request'

const namespaceName = 'callpathtest'

describe('Test subflow behaviour', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	it(`should create a namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }`)

		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			namespace: {
				createdAt: expect.stringMatching(common.regex.timestampRegex),
				updatedAt: expect.stringMatching(common.regex.timestampRegex),
				name: namespaceName,
			},
		})
	})

	it(`should create a directory called /a`, async () => {
		const createDirectoryResponse = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }/tree/a?op=create-directory`)
		expect(createDirectoryResponse.statusCode).toEqual(200)
	})

	it(`should create a workflow called /a/child.yaml`, async () => {
		const res = await request(common.config.getDirektivHost())
			.put(`/api/namespaces/${ namespaceName }/tree/a/child.yaml?op=create-workflow`)
			.set({
				'Content-Type': 'text/plain',
			})
			.send(`
states:
- id: a
  type: noop
  transform:
    result: 'jq(.input + 1)'`)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: namespaceName,
		})
	})

	it(`should create a workflow called /a/parent1.yaml`, async () => {
		const res = await request(common.config.getDirektivHost())
			.put(`/api/namespaces/${ namespaceName }/tree/a/parent1.yaml?op=create-workflow`)
			.set({
				'Content-Type': 'text/plain',
			})
			.send(`
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
`)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: namespaceName,
		})
	})

	it(`should invoke the '/a/parent1.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/a/parent1.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			result: 2,
		})
	})

	it(`check if child logs are present in parent's log view`, async () => {
		const instances = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/instances`)
		expect(instances.statusCode).toEqual(200)
		expect(instances.body.instances.results.length).not.toBeLessThan(1)
	})

	// it(`check if parentslogs contain child logs`, async () => {
	//     var instancesSource = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances?filter.field=AS&filter.type=WORKFLOW&filter.val=a/parent1.yaml`)
	//     var instancesChild = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances?filter.field=AS&filter.type=WORKFLOW&filter.val=/a/child.yaml`)
	//     const idsS = instancesSource.body.instances.results.map((result) => result.id)
	//     const idsC = instancesChild.body.instances.results.map((result) => result.id)
	//     const logsSourceResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances/${idsS}/logs`)
	//     expect(logsSourceResponse.statusCode).toEqual(200)
	//     const logsChildResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances/${idsC}/logs`)
	//     expect(logsChildResponse.statusCode).toEqual(200)
	//     expect(logsChildResponse.body.results.length).toBeLessThan(logsSourceResponse.body.results.length)
	//     expect(logsChildResponse.body.results.length).not.toBeLessThan(1)
	// })

	// it(`check if the logs are structured properly`, async () => {
	//     var instancesSource = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances?filter.field=AS&filter.type=WORKFLOW&filter.val=a/parent1.yaml`)
	//     var instancesChild = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances?filter.field=AS&filter.type=WORKFLOW&filter.val=/a/child.yaml`)
	//     const idsS = instancesSource.body.instances.results.map((result) => result.id)
	//     const idsC = instancesChild.body.instances.results.map((result) => result.id)
	//     const logsSourceResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances/${idsS}/logs`)
	//     expect(logsSourceResponse.statusCode).toEqual(200)
	//     await Promise.all(logsSourceResponse.body.results.map(async (logEntry) => {
	//         expect(logEntry["tags"]["callpath"]==`/${idsS}` || logEntry["tags"]["callpath"] == `/${idsS}/${idsC}`).toBeTruthy()
	//         expect(logEntry["tags"]["namespace-id"]).toMatch(/^.{8}-.{4}-.{4}-.{4}-.{12}$/)
	//         expect(logEntry["tags"]["workflow"]==`a/parent1.yaml` || logEntry["tags"]["workflow"] == `/a/child.yaml`).toBeTruthy()
	//         expect(logEntry["tags"]["trace"]).toMatch(/^.{32}$/)
	//         expect(logEntry["tags"]["level"]).toMatch(/^(info|debug|error)$/)
	//         expect(logEntry["tags"]["revision-id"]).toMatch(/^.{8}-.{4}-.{4}-.{4}-.{12}$/)
	//         expect(logEntry["tags"]["instance-id"]==idsS || logEntry["tags"]["instance-id"] == idsC).toBeTruthy()
	//         expect(logEntry["tags"]["root-instance-id"]).toEqual(`${idsS}`)
	//     }))
	//     const logsChildResponse = await request(common.config.getDirektivHost()).get(`/api/namespaces/${namespaceName}/instances/${idsC}/logs`)
	//     expect(logsChildResponse.statusCode).toEqual(200)
	//     expect(logsChildResponse.body.results.length).toBeLessThan(logsSourceResponse.body.results.length)
	//     expect(logsChildResponse.body.results.length).not.toBeLessThan(1)
	//     await Promise.all(logsChildResponse.body.results.map(async (logEntry) => {
	//         expect(logEntry["tags"]["callpath"]).toEqual(`/${idsS}/${idsC}`)
	//         expect(logEntry["tags"]["workflow"]).toEqual(`/a/child.yaml`)
	//         expect(logEntry["tags"]["revision-id"]).toMatch(/^.{8}-.{4}-.{4}-.{4}-.{12}$/)
	//     }))
	// })
})
