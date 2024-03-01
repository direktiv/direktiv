import { beforeAll, describe, expect, it } from '@jest/globals'

import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

describe('Test namespace create calls', () => {
	beforeAll(helpers.deleteAllNamespaces)

	const testCases = [
		{
			input: {
				name: 'foo1',
			},
			want: {
				name: 'foo1',
			},
		},
		{
			input: {
				name: 'foo2',
				mirrorSettings: {
					url: 'my_url',
				},
			},
			want: {
				name: 'foo2',
				mirrorSettings: {
					url: 'my_url',
					gitCommitHash: '',
					gitRef: '',
					insecure: false,
					publicKey: '',
					createdAt: expect.stringMatching(regex.timestampRegex),
					updatedAt: expect.stringMatching(regex.timestampRegex),
				},
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		it(`should create a new namespace case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.post(`/api/v2/namespaces`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(200)
			expect(res.body.data).toEqual({
				...testCase.want,

				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
			})
		})
	}
})

describe('Test invalid namespace create calls', () => {
	beforeAll(helpers.deleteAllNamespaces)

	const testCases = [
		{
			input: {
				// invalid data
				name: '11',
			},
			wantError: {
				statusCode: 400,
				error: {
					code: 'request_data_invalid',
					message: 'invalid namespace name',
				},
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		it(`should fail create a new namespace case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.post(`/api/v2/namespaces`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(testCase.wantError.statusCode)
			expect(res.body.error).toEqual(
				testCase.wantError.error,
			)
		})
	}
})
