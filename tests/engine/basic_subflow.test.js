import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import { fileURLToPath } from 'url'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

function randomLowercaseString(length) {
	const letters = 'abcdefghijklmnopqrstuvwxyz'
	let result = ''
	for (let i = 0; i < length; i++) {
		result += letters[Math.floor(Math.random() * 26)]
	}
	return result
}

const namespace =
	randomLowercaseString(3) + '-' + basename(fileURLToPath(import.meta.url))

describe('Test js engine', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldTSWorkflow(
		it,
		expect,
		namespace,
		'/',
		'sub.wf.ts',
		`
function stateOne(payload) {
	print("(SUBFLOW) RUN STATE ONE");
	payload.subflow1 = "OK1";	
	
	return transition(stateTwo, payload);
}
function stateTwo(payload) {
	print("(SUBFLOW) RUN STATE TWO");
	payload.subflow2 = "OK2";	
	
    return finish(payload);
}
		`,
	)

	helpers.itShouldTSWorkflow(
		it,
		expect,
		namespace,
		'/',
		'main.wf.ts',
		`
function stateOne(payload) {
	print("(MAIN) RUN STATE ONE");
	payload.main1 = "OK1";	
	
	return transition(stateTwo, payload);
}
function stateTwo(payload) {
	print("(MAIN) RUN STATE TWO");
	payload.main2 = "OK2";	
	
    return finish(payload);
}
		`,
	)

	it(`should invoke /sub.wf.ts workflow`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.post(
				`/api/v2/namespaces/${namespace}/instances?path=/sub.wf.ts&wait=true&fullOutput=true`,
			)
			.send({ foo: 'bar' })
		console.log(res.body)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.status).toEqual('complete')
		expect(res.body.data.output).toEqual(
			JSON.stringify({ foo: 'bar', subflow1: 'OK1', subflow2: 'OK2' }),
		)
	})

	it(`should invoke /main.wf.ts workflow`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.post(
				`/api/v2/namespaces/${namespace}/instances?path=/main.wf.ts&wait=true&fullOutput=true`,
			)
			.send({ foo: 'bar' })
		console.log(res.body)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.status).toEqual('complete')
		expect(res.body.data.output).toEqual(
			JSON.stringify({ foo: 'bar', main1: 'OK1', main2: 'OK2' }),
		)
	})
})
