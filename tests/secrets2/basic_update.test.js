import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import request from "../common/request"

import common from '../common'
import config from '../common/config'
import helpers from '../common/helpers'
import regex from '../common/regex'

const namespace = basename(__filename)

describe('Test secret update calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	let createRes
	it(`should create a secret case`, async () => {
		createRes = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/secrets`)
			.send({
				name: 'foo',
				data: btoa('bar'),
				
			})

		expect(createRes.statusCode).toEqual(200)
	})

	const testCases = [
		{
			input: {
				data: btoa('bar2--'),
			},
			want: {
				name: 'foo',
			},
		},
	]

	for (let i = 0; i < testCases.length; i++) {
		const testCase = testCases[i]

		it(`should update secret case ${ i }`, async () => {
			const secretName = createRes.body.data.name
			const res = await request(config.getDirektivHost())
				.patch(`/api/v2/namespaces/${ namespace }/secrets/${ secretName }`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(200)
			expect(res.body.data).toMatchObject({
				...testCase.want,

				createdAt: expect.stringMatching(regex.timestampRegex),
				updatedAt: expect.stringMatching(regex.timestampRegex),
			})
		})
	}
})

describe('Test invalid secret update calls', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	let createRes
	it(`should create a secret case`, async () => {
		createRes = await request(config.getDirektivHost())
			.post(`/api/v2/namespaces/${ namespace }/secrets`)
			.send({
				name: 'foo',
				data: btoa('bar'),
				
			})

		expect(createRes.statusCode).toEqual(200)
	})

	const testCases = [
		{
			input: {
				// invalid data
				data: 'some string',
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

		it(`should fail updating a secret case ${ i }`, async () => {
			const secretName = createRes.body.data.name
			const res = await request(config.getDirektivHost())
				.patch(`/api/v2/namespaces/${ namespace }/secrets/${ secretName }`)
				.send(testCase.input)
			expect(res.statusCode).toEqual(testCase.wantError.statusCode)
			expect(res.body.error).toMatchObject(
				testCase.wantError.error,
			)
		})
	}
})
