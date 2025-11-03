import { beforeAll, describe, expect, it } from '@jest/globals'
import { btoa } from 'js-base64'
import { basename } from 'path'
import { fileURLToPath } from 'url'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test js engine', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	const testCases = [
		{
			name: 'singleStep.wf.ts',
			input: { foo: 'bar' },
			wantOutput: JSON.stringify('done'),
			wantErrorMessage: null,
			wantStatus: 'complete',
			file: `
function stateOne(payload) {
	return finish("done");
}`,
		},
		{
			name: 'singleStep2.wf.ts',
			input: { foo: 'bar' },
			wantOutput: JSON.stringify('done'),
			wantErrorMessage: null,
			wantStatus: 'complete',
			file: `
function stateOne(payload) {
	return finish("done");
}`,
		},
		{
			name: 'throwError.wf.ts',
			input: JSON.stringify('anything'),
			wantOutput: null,
			wantErrorMessage:
				'invoke start: simply failed at stateOne (throwError.wf.ts:3:1(2))',
			wantStatus: 'failed',
			file: `
function stateOne(payload) {
	throw "simply failed";
	return finish("was ok");
}`,
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]
		helpers.itShouldCreateFile(
			it,
			expect,
			namespace,
			'/',
			testCase.name,
			'workflow',
			'application/x-typescript',
			btoa(testCase.file),
		)
		it(`should invoke /${testCase.name} workflow`, async () => {
			const res = await request(common.config.getDirektivBaseUrl())
				.post(
					`/api/v2/namespaces/${namespace}/instances?path=/${testCase.name}&wait=true`,
				)
				.send(testCase.input)
			expect(res.statusCode).toEqual(200)
		})
	}

	it(`should filter instances by filter[status]=complete`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.get(
				`/api/v2/namespaces/${namespace}/instances?filter[status]=complete`,
			)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.length).toBe(2)
	})
	it(`should filter instances by filter[status]=failed`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.get(
				`/api/v2/namespaces/${namespace}/instances?filter[status]=failed`,
			)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.length).toBe(1)
	})
	it(`should list all instances`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.get(
				`/api/v2/namespaces/${namespace}/instances`,
			)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.length).toBe(3)
	})
})
