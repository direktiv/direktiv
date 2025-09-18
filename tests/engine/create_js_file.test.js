import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import helpers from '../common/helpers'
import request from "../common/request";
import common from "../common";
import {retry10} from "../common/retry";


const namespace = basename(__filename.replaceAll('.', '-'));
describe('Test js engine', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)
	helpers.itShouldCreateDir(it, expect, namespace, '/', 'foo')

	helpers.itShouldCreateFile(it, expect, namespace, '/foo', 'file1.wf.js', 'file', 'text/plain',
		btoa(`
function stateOne(payload) {
	print("RUN STATE FIRST")
    return transition("stateTwo", payload)
}
function stateTwo(payload) {
	print("RUN STATE FIRST")
	return finish("hello")
}

`))
	retry10(`should invoke /foo/file1.wf.js workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/instances?path=foo/file1.wf.js`).send({ foo: "bar" })
		console.log(req.text)
		expect(req.statusCode).toEqual(200)
	})

})
