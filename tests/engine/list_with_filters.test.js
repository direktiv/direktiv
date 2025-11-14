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
					`/api/v2/namespaces/${namespace}/instances?path=/${testCase.name}&wait=true&fullOutput=true`,
				)
				.send(testCase.input)
			expect(res.statusCode).toEqual(200)
		})
	}

	const filterCases = [
		{
			query: '?filter[status]=complete',
			wantCount: 2,
			wantStatuses: ['complete', 'complete'],
		},
		{
			query: '?filter[status][eq]=complete',
			wantCount: 2,
			wantStatuses: ['complete', 'complete'],
		},
		{
			query: '?filter[status][in]=complete',
			wantCount: 2,
			wantStatuses: ['complete', 'complete'],
		},
		{
			query: '?filter[status][cn]=comp',
			wantCount: 2,
			wantStatuses: ['complete', 'complete'],
		},
		{
			query: '?filter[status]=failed',
			wantCount: 1,
			wantStatuses: ['failed'],
		},
		{
			query: '?filter[status][eq]=failed',
			wantCount: 1,
			wantStatuses: ['failed'],
		},
		{
			query: '?filter[status][in]=failed',
			wantCount: 1,
			wantStatuses: ['failed'],
		},
		{
			query: '?filter[status][cn]=fail',
			wantCount: 1,
			wantStatuses: ['failed'],
		},
		{
			query: '?filter[status][in]=complete,failed',
			wantCount: 3,
			wantStatuses: ['failed', 'complete', 'complete'],
		},
		{
			query: '',
			wantCount: 3,
			wantStatuses: ['failed', 'complete', 'complete'],
		},
		{
			query: '?filter[status][cn]=le',
			wantCount: 3,
			wantStatuses: ['failed', 'complete', 'complete'],
		},
		{
			query: '?filter[status]=nothing',
			wantCount: 0,
			wantStatuses: [],
		},
	]

	for (let i = 0; i < filterCases.length; i++) {
		const filterCase = filterCases[i]
		it(`should list instances with filter ${filterCase.query}`, async () => {
			const res = await request(common.config.getDirektivBaseUrl()).get(
				`/api/v2/namespaces/${namespace}/instances${filterCase.query}`,
			)
			expect(res.statusCode).toEqual(200)
			expect(res.body.data.length).toBe(filterCase.wantCount)
			expect(res.body.data.length).toBe(filterCase.wantStatuses.length)
			expect(res.body.data.map((i) => i.status)).toEqual(
				filterCase.wantStatuses,
			)
		})
	}
})
