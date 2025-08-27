import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import helpers from '../common/helpers'
import request from "../common/request";
import common from "../common";


const namespace = basename(__filename.replaceAll('.', '-'));
describe('Test js engine', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)
	helpers.itShouldCreateDir(it, expect, namespace, '/', 'foo')

	helpers.itShouldCreateFile(it, expect, namespace, '/foo', 'file1.wf.js', 'file', 'text/plain',
		btoa(`
function start(input) {
	print("RUN STATE FIRST")
	state = input
	state.step1 = "data1"
	
	return transition(second, state)
}
function second(state) {
    print("RUN STATE SECOND")
    state.step2 = "data2"

    return transition(third, state)
}
function third(state) {
    print("RUN STATE LAST")
    state.step3 = "data3"
    
    print(state)
	return state
}
`))
	it(`should invoke /foo/file1.wf.js workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/instances?path=foo/file1.wf.js`).send({ foo: "bar" })
		console.log(req.text)
		expect(req.statusCode).toEqual(200)
	})

})
