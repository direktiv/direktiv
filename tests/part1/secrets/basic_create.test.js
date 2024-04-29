import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../../common/config'
import helpers from '../../common/helpers'
import regex from '../../common/regex'
import request from '../../common/request'

const namespace = basename(__filename)

describe('Test secret create calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	const testCases = [
		{
			input: {
				name: 'foo1',
				data: btoa('bar1'),
			},
			want: {
				name: 'foo1',
				initialized: true,
			},
		},
		{
			input: {
				name: 'foo2',
				data: btoa('bar2'),
			},
			want: {
				name: 'foo2',
				initialized: true,
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		it(`should create a new secret case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.post(`/api/v2/namespaces/${ namespace }/secrets`)
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

describe('Test invalid secret create calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	const testCases = [
		{
			input: {
				// invalid data
				name: 'foo1',
				data: 'invalid-base-64',

			},
			wantError: {
				statusCode: 400,
				error: {
					code: 'request_body_not_json',
					message: "couldn't parse request payload in json format",
				},
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		it(`should fail create a new secret case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.post(`/api/v2/namespaces/${ namespace }/secrets`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(testCase.wantError.statusCode)
			expect(res.body.error).toEqual(
				testCase.wantError.error,
			)
		})
	}
})
