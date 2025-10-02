import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const namespace = basename(__filename.replaceAll('.', '-'))

describe('Test js engine', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)
	helpers.itShouldCreateDir(it, expect, namespace, '/', 'foo')

	const testCases = [
		{ name: 'singleStep.wf.ts',
			input: { foo: 'bar' },
			wantOutput: 'done',
			wantErrorMessage: null,
			file: `
function stateOne(payload) {
	return finish("done");
}`		},
		{ name: 'twoSteps.wf.ts',
			input: { foo: 'bar' },
			wantOutput: { foo: 'bar', bar: 'foo' },
			wantErrorMessage: null,
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
		helpers.itShouldCreateFile(it, expect, namespace, '/foo', testCase.name, 'workflow', 'application/x-typescript',
			btoa(testCase.file))
		retry10(`should invoke /foo/${ testCase.name } workflow`, async () => {
			const res = await request(common.config.getDirektivBaseUrl()).post(`/api/v2/namespaces/${ namespace }/instances?path=foo/${ testCase.name }`)
				.send(testCase.input)
			console.log(res.statusCode, res.text)
			expect(res.statusCode).toEqual(200)
			expect(res.body.data.errorMessage).toEqual(testCase.wantErrorMessage)
			let gotOutput = atob(res.body.data.output)
			gotOutput = JSON.parse(gotOutput)
			expect(gotOutput).toEqual(testCase.wantOutput)
		})
	}
})
