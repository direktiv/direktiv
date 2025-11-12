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
			wantOutput: {"data":"done"},
			statusCode: 200,
			file: `
function stateOne(payload) {
	return finish("done");
}`,
		},
		{
			name: 'twoSteps.wf.ts',
			input: JSON.stringify({ foo: 'bar' }),
			wantOutput:  {"data":{"bar":"foo","foo":"bar"}},
			statusCode: 200,
			file: `
function stateOne(payload) {
	print("RUN STATE FIRST");
	payload.bar = "foo";
	return transition(stateTwo, payload);
}
function stateTwo(payload) {
	print("RUN STATE SECOND");
    return finish(payload);
}`,
		},
		{
			name: 'stringInput.wf.ts',
			input: JSON.stringify('hello'),
			wantOutput: {"data":"helloWorld"},
			statusCode: 200,
			file: `
function stateOne(payload) {
	return finish(payload + "World");
}`,
		},
		{
			name: 'numberInput.wf.ts',
			input: JSON.stringify(146),
			wantOutput:  {"data":147},
			statusCode: 200,
			file: `
function stateOne(payload) {
	return finish(payload + 1);
}`,
		},
		{
			name: 'throwError.wf.ts',
			input: JSON.stringify('anything'),
			wantOutput: {"error":{"code":"","message":"invoke start: simply failed at stateOne (throwError.wf.ts:3:1(2))"}},
			statusCode: 500,
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
		it(`should invoke /${testCase.name} workflow with &wait=true`, async () => {
			const res = await request(common.config.getDirektivBaseUrl())
				.post(
					`/api/v2/namespaces/${namespace}/instances?path=/${testCase.name}&wait=true`,
				)
				.send(testCase.input)
			expect(res.statusCode).toEqual(testCase.statusCode)
		})
	}
})
