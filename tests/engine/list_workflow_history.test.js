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
	payload.one = 1;
	return transition(stateTwo, payload);
}

function stateTwo(payload) {
	print("RUN STATE SECOND");
	payload.two = 2;
    return transition(stateThree, payload);
}

function stateThree(payload) {
	print("RUN STATE THIRD");
	payload.three = 3;
    return finish(payload);
}` },
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
			instanceId = res.body.data.id
		})

		it(`should list /${ testCase.name } workflow history`, async () => {
			const res = await request(common.config.getDirektivBaseUrl()).get(`/api/v2/namespaces/${ namespace }/instances/${ instanceId }/history`)
			expect(res.statusCode).toEqual(200)
			const history = res.body.data.map(item => ({ type: item.type,
				fn: item.fn,
				input: item.input,
				memory: item.memory,
				output: item.output,
				sequence: item.sequence }))

			let firstSequence = history[0].sequence
			expect(history).toEqual([
				{
					type: 'pending',
					fn: 'stateOne',
					input: { foo: 'bar' },
					memory: undefined,
					output: undefined,
					sequence: firstSequence++,
				},
				{
					type: 'running',
					fn: 'stateOne',
					input: { foo: 'bar' },
					memory: undefined,
					output: undefined,
					sequence: firstSequence++,
				},
				{
					type: 'running',
					fn: 'stateTwo',
					input: undefined,
					memory: { foo: 'bar', one: 1 },
					output: undefined,
					sequence: firstSequence++,
				},
				{
					type: 'running',
					fn: 'stateThree',
					input: undefined,
					memory: { foo: 'bar', one: 1, two: 2 },
					output: undefined,
					sequence: firstSequence++,
				},
				{
					type: 'succeeded',
					fn: undefined,
					input: undefined,
					memory: undefined,
					output: { foo: 'bar', one: 1, three: 3, two: 2 },
					sequence: firstSequence++,
				},
			])
		})
	}
})
