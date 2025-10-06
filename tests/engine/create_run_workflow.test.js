import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import {btoa} from "js-base64";

const namespace = basename(__filename.replaceAll('.', '-'))

describe('Test js engine', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	const testCases = [
		{ name: 'singleStep.wf.ts',
			input: { foo: 'bar' },
			wantOutput: JSON.stringify('done'),
			wantErrorMessage: null,
			wantStatus: 'complete',
			file: `
function stateOne(payload) {
	return finish("done");
}`		},
		{ name: 'twoSteps.wf.ts',
			input: JSON.stringify({ foo: 'bar' }),
			wantOutput: JSON.stringify({ bar: 'foo', foo: 'bar' }),
			wantErrorMessage: null,
			wantStatus: 'complete',
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
		{ name: 'stringInput.wf.ts',
			input: JSON.stringify("hello"),
			wantOutput: JSON.stringify('helloWorld'),
			wantErrorMessage: null,
			wantStatus: 'complete',
			file: `
function stateOne(payload) {
	return finish(payload + "World");
}`		},
		{ name: 'numberInput.wf.ts',
			input: JSON.stringify(146),
			wantOutput: JSON.stringify(147),
			wantErrorMessage: null,
			wantStatus: 'complete',
			file: `
function stateOne(payload) {
	return finish(payload + 1);
}`		},
		{ name: 'throwError.wf.ts',
			input: JSON.stringify("anything"),
			wantOutput: null,
			wantErrorMessage: btoa("invoke start: simply failed at stateOne (throwError.wf.ts:3:1(2))"),
			wantStatus: 'failed',
			file: `
function stateOne(payload) {
	throw "simply failed";
	return finish("was ok");
}`		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]
		helpers.itShouldCreateFile(it, expect, namespace, '/', testCase.name, 'workflow', 'application/x-typescript',
			btoa(testCase.file))
		it(`should invoke /${ testCase.name } workflow`, async () => {
			const res = await request(common.config.getDirektivBaseUrl()).post(`/api/v2/namespaces/${ namespace }/instances?path=/${ testCase.name }`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(200)
			expect(res.body.data.status).toEqual(testCase.wantStatus)
			expect(res.body.data.errorMessage).toEqual(testCase.wantErrorMessage)
			expect(res.body.data.output).toEqual(testCase.wantOutput)
		})
	}
})
