import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

describe('Test namespace update calls', () => {
	beforeAll(helpers.deleteAllNamespaces)

	let createRes
	it(`should create a namespace case`, async () => {
		createRes = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces`)
			.send({
				name: 'foo',
			})

		expect(createRes.statusCode).toEqual(200)
	})

	const testCases = [
		{
			input: {
				mirrorSettings: {
					url: "my_url",
				}
			},
			want: {
				name: 'foo',
				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
				mirrorSettings: {
					url: "my_url",
					gitCommitHash: "",
					gitRef: "",
					insecure: false,
					publicKey: "",
					createdAt: expect.stringMatching(regex.timestampRegex),
					updatedAt: expect.stringMatching(regex.timestampRegex),
				}
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		// eslint-disable-next-line no-loop-func
		it(`should update namespace case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.put(`/api/v2/namespaces/${ createRes.body.data.name }`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(200)
			expect(res.body.data).toEqual({
				...testCase.want,
			})
		})
	}
})

describe('Test invalid namespace update calls', () => {
	beforeAll(helpers.deleteAllNamespaces)

	let createRes
	it(`should create a namespace case`, async () => {
		createRes = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces`)
			.send({
				name: 'foo',
				data: btoa('bar'),

			})

		expect(createRes.statusCode).toEqual(200)
	})

	const testCases = [
		{
			input: {
				mirrorSettings: {
					url: "my_url_invalid",
				}
			},
			wantError: {
				statusCode: 400,
				error: {
					code: 'request_body_invalid',
					message: "invalid url",
				},
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]
		// eslint-disable-next-line no-loop-func
		it(`should fail updating a namespace case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.put(`/api/v2/namespaces/${ createRes.body.data.name }`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(testCase.wantError.statusCode)
			expect(res.body.error).toEqual(
				testCase.wantError.error,
			)
		})
	}
})
