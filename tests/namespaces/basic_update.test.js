import { beforeAll, describe, expect, it } from '@jest/globals'

import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'
import request from '../common/request'

const timestamps = {
	createdAt: expect.stringMatching(regex.timestampRegex),
	updatedAt: expect.stringMatching(regex.timestampRegex),
}

describe('Test namespace simple update calls', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, 'foo')

	const testCases = [
		{
			input: {
				mirror: {
					url: 'my_url',
					gitRef: 'main',
					authType: 'public',
				},
			},
			want: {
				name: 'foo',
				isSystemNamespace: false,
				...timestamps,
				mirror: {
					url: 'my_url',
					gitRef: 'main',
					insecure: false,
					authType: 'public',
					...timestamps,
				},
			},
		},
		{
			input: {
				mirror: {
					url: 'my_url2',
					insecure: true,
					gitRef: 'master',
					authType: 'public',
				},
			},
			want: {
				name: 'foo',
				isSystemNamespace: false,
				...timestamps,
				mirror: {
					url: 'my_url2',
					insecure: true,
					gitRef: 'master',
					authType: 'public',
					...timestamps,
				},
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		// eslint-disable-next-line no-loop-func
		it(`should update namespace case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.patch(`/api/v2/namespaces/foo`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(200)
			expect(res.body.data).toEqual({
				...testCase.want,
			})
		})
	}
})

describe('Test namespace mirror update calls', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, 'foo')

	const testCases = [
		{
			input: {
				url: 'my_url',
				gitRef: 'main',
				authType: 'public',
			},
			want: {
				url: 'my_url',
				gitRef: 'main',
				authType: 'public',
				insecure: false,
			},
		},
		{
			input: {
				url: 'my_url2',
			},
			want: {
				url: 'my_url2',
				gitRef: 'main',
				authType: 'public',
				insecure: false,
			},
		},
		{
			input: {
				gitRef: 'main2',
			},
			want: {
				url: 'my_url2',
				gitRef: 'main2',
				authType: 'public',
				insecure: false,
			},
		},
		{
			input: {
				gitRef: 'main2',
				authType: 'token',
				authToken: '1234',
			},
			want: {
				url: 'my_url2',
				gitRef: 'main2',
				authType: 'token',
				insecure: false,
			},
		},
		{
			input: {
				gitRef: 'main2',
				authType: 'public',
			},
			want: {
				url: 'my_url2',
				gitRef: 'main2',
				authType: 'public',
				insecure: false,
			},
		},
		{
			input: {
				gitRef: 'main2',
				authType: 'ssh',
				publicKey: 'my-ppk',
				privateKey: 'my-pvk',
			},
			want: {
				url: 'my_url2',
				gitRef: 'main2',
				authType: 'ssh',
				publicKey: 'my-ppk',
				insecure: false,
			},
		},
		{
			input: {
				gitRef: 'main2',
				authType: 'public',
			},
			want: {
				url: 'my_url2',
				gitRef: 'main2',
				authType: 'public',
				insecure: false,
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		// eslint-disable-next-line no-loop-func
		it(`should update namespace case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.patch(`/api/v2/namespaces/foo`)
				.send({ mirror: testCase.input })
			expect(res.statusCode).toEqual(200)
			expect(res.body.data).toEqual({
				name: 'foo',
				isSystemNamespace: false,
				...timestamps,
				mirror: {
					...testCase.want,
					...timestamps,
				},
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
			})

		expect(createRes.statusCode).toEqual(200)
	})

	const testCases = [
		{
			input: {
				mirror: {
					url: 11,
				},
			},
			wantError: {
				statusCode: 400,
				error: {
					code: 'request_body_bad_json_schema',
					message: 'request payload has bad json schema',
				},
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]
		// eslint-disable-next-line no-loop-func
		it(`should fail updating a namespace case ${ i }`, async () => {
			const res = await request(config.getDirektivHost())
				.patch(`/api/v2/namespaces/${ createRes.body.data.name }`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(testCase.wantError.statusCode)
			expect(res.body.error).toEqual(
				testCase.wantError.error,
			)
		})
	}
})
