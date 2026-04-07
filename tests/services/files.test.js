import { afterEach, beforeEach, describe, expect, it } from '@jest/globals'
import { bashServiceSrc, workflowFilesInSystemServiceSrc } from './fixtures'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { waitForInstanceStatus } from '../instances/utils'
import { waitForServiceCondition } from './utils'

const baseUrl = config.getDirektivBaseUrl()
const systemNamespace = 'system'
let namespace = ''

describe('Files mounted into execService', () => {
	beforeEach(async () => {
		namespace = helpers.randomNamespaceName()

		const systemNsResponse = await request(baseUrl)
			.post('/api/v2/namespaces')
			.send({ name: systemNamespace })
		expect(systemNsResponse.statusCode).toEqual(200)

		const nsResponse = await request(baseUrl)
			.post('/api/v2/namespaces')
			.send({ name: namespace })
		expect(nsResponse.statusCode).toEqual(200)
	})

	afterEach(async () => {
		return Promise.all([
			helpers.deleteNamespace(systemNamespace),
			helpers.deleteNamespace(namespace),
		])
	})

	it('mounts namespace variables and namespace files into an execService run', async () => {
		// create system service
		const systemServiceFileName = 'bash.svc.json'
		const systemServiceResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${systemNamespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: systemServiceFileName,
				type: 'service',
				mimeType: 'application/json',
				data: btoa(bashServiceSrc),
			})
		expect(systemServiceResponse.statusCode).toEqual(200)

		// confirm system service is running
		const expectedCondition = {
			type: 'Available',
			status: 'True',
		}
		const service = await waitForServiceCondition(
			systemNamespace,
			`/${systemServiceFileName}`,
			expectedCondition,
		)
		expect(service).toBeDefined()

		// create workflow
		const workflowResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'files-in-service.wf.ts',
				type: 'workflow',
				mimeType: 'application/typescript',
				data: btoa(workflowFilesInSystemServiceSrc),
			})
		expect(workflowResponse.statusCode).toEqual(200)

		// set up namespace file to be mounted via scope "file"
		const fileResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/files`)
			.set('Content-Type', 'application/json')
			.send({
				name: 'test.wf.ts',
				type: 'file',
				mimeType: 'application/typescript',
				data: btoa(`const flow: FlowDefinition = { type: "default" };`),
			})
		expect(fileResponse.statusCode).toEqual(200)

		// set up namespace variable to be mounted via scope "namespace"
		const varResponse = await request(baseUrl)
			.post(`/api/v2/namespaces/${namespace}/variables`)
			.send({
				name: 'foobar',
				data: btoa('proof-foobar'),
				mimeType: 'text/plain',
			})
		expect(varResponse.statusCode).toEqual(200)

		// execute workflow
		const executeResponse = await request(baseUrl).post(
			`/api/v2/namespaces/${namespace}/instances?path=files-in-service.wf.ts`,
		)
		expect(executeResponse.statusCode).toEqual(200)

		const { id } = executeResponse.body.data
		const instance = await waitForInstanceStatus(
			baseUrl,
			namespace,
			id,
			'complete',
		)
		expect(instance).toBeDefined()
		expect(instance.body.data.status).toEqual('complete')

		const output = JSON.parse(instance.body.data.output)

		// files are listed with ls command (including permissions for 0755);
		// namespace variables are mounted as files too (e.g. foobar)
		expect(output?.bash?.[0]?.success).toEqual(true)
		expect(output?.bash?.[0]?.result).toEqual(
			expect.stringContaining('test.wf.ts'),
		)
		expect(output?.bash?.[0]?.result).toEqual(
			expect.stringContaining('foobar'),
		)
		expect(output?.bash?.[0]?.result).toEqual(
			expect.stringMatching(/-rw-r--r--.*\stest\.wf\.ts$/m),
		)
		expect(output?.bash?.[0]?.result).toEqual(
			expect.stringMatching(/-rwxr-xr-x.*\sfoobar$/m),
		)

		// mounted namespace variable content is readable
		expect(output?.bash?.[1]).toMatchObject({
			success: true,
			result: 'proof-foobar',
		})

		// mounted file content is readable
		expect(output?.bash?.[2]?.success).toEqual(true)
		expect(output?.bash?.[2]?.result).toEqual(
			expect.stringContaining('const flow: FlowDefinition'),
		)
	})
})
