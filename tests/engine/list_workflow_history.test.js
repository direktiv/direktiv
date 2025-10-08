import { beforeAll, describe, expect, it } from '@jest/globals'
import { btoa } from 'js-base64'
import { basename } from 'path'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespace = basename(__filename)

describe('List workflow history', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	const testCases = [
		{ name: 'twoSteps.wf.ts',
			input: JSON.stringify({ foo: 'bar' }),
			file: `
function stateOne(payload) {
	print("RUN STATE FIRST");
	payload.bar = "foo";
	return transition(stateTwo, payload);
}
function stateTwo(payload) {
	print("RUN STATE SECOND");
    return finish(payload);
}`		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]
		helpers.itShouldCreateFile(it, expect, namespace, '/', testCase.name, 'workflow', 'application/x-typescript',
			btoa(testCase.file))
		let instanceId = null

		it(`should invoke /${ testCase.name } workflow`, async () => {
			const res = await request(common.config.getDirektivBaseUrl()).post(`/api/v2/namespaces/${ namespace }/instances?path=/${ testCase.name }&wait=true`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(200)
			console.log(res.body.data.id)
			instanceId = res.body.data.id
		})

		it(`should list /${ testCase.name } workflow history`, async () => {
			console.log(instanceId)
			const res = await request(common.config.getDirektivBaseUrl()).get(`/api/v2/namespaces/${ namespace }/instances/${ instanceId }/history`)
			expect(res.statusCode).toEqual(200)
			console.log(res.body.data)
		})
	}
})
