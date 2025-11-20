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
	payload.events += "_sub1";

	return transition(stateTwo, payload);
}
function stateTwo(payload) {
	print("(SUBFLOW) RUN STATE TWO");
	payload.events += "_sub2";
	
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
	payload.events += "_start";	
	
	output = execSubflow("/sub.wf.ts", "stateOne" ,payload);
	
	return transition(stateTwo, output);
}
function stateTwo(payload) {
	print("(MAIN) RUN STATE TWO");
	payload.events += "_end";
	
    return finish(payload);
}
		`,
	)

	it(`should invoke /sub.wf.ts workflow`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.post(
				`/api/v2/namespaces/${namespace}/instances?path=/sub.wf.ts&wait=true&fullOutput=true`,
			)
			.send({ events: '' })
		console.log(res.body)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.status).toEqual('complete')
		expect(res.body.data.output).toEqual(
			JSON.stringify({ events: '_sub1_sub2' }),
		)
	})

	it(`should invoke /main.wf.ts workflow`, async () => {
		const res = await request(common.config.getDirektivBaseUrl())
			.post(
				`/api/v2/namespaces/${namespace}/instances?path=/main.wf.ts&wait=true&fullOutput=true`,
			)
			.send({ events: '' })
		console.log(res.body)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.status).toEqual('complete')
		expect(res.body.data.output).toEqual(
			JSON.stringify({ events: '_start_sub1_sub2_end' }),
		)
	})
})
