import request from 'supertest'

import common from '../common'

const testNamespace = 'test-file-namespace'

beforeAll(async () => {
	// delete a 'test-namespace' if it's already exit.
	await request(common.config.getDirektivHost()).delete(`/api/namespaces/${ testNamespace }?recursive=true`)
})

describe('Test namespaces crud operations', () => {
	common.helpers.itShouldCreateNamespace(it, expect, testNamespace)

	it(`should create a new direktiv file`, async () => {

		const res = await request(common.config.getDirektivHost())
			.put(`/api/namespaces/${ testNamespace }/tree/my-workflow.yaml?op=create-workflow`)
			.set({
				'Content-Type': 'text/plain',
			})

			.send(`
description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!`)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: testNamespace,
		})
	})
})


// Current response structure:
// {
//    "namespace": "test-namespace",
//    "node": {
//    "createdAt": "2023-03-01T16:38:15.177306157Z",
//        "updatedAt": "2023-03-01T16:38:15.177306608Z",
//        "name": "ddd",
//        "path": "/ddd",
//        "parent": "/",
//        "type": "workflow",
//        "attributes": [],
//        "readOnly": false,
//        "expandedType": "workflow"
// },
//    "revision": {
//    "createdAt": "2023-03-01T16:38:15.178865993Z",
//        "hash": "0d2cade3a4196e41b07524d747df3ef54e73f2735f8e25c74e1ecbf9498f8dff",
//        "source": "ZGVzY3JpcHRpb246IEEgc2ltcGxlICduby1vcCcgc3RhdGUgdGhhdCByZXR1cm5zICdIZWxsbyB3b3JsZCEnCnN0YXRlczoKLSBpZDogaGVsbG93b3JsZAogIHR5cGU6IG5vb3AKICB0cmFuc2Zvcm06CiAgICByZXN1bHQ6IEhlbGxvIHdvcmxkIQo=",
//        "name": "714fe9c9-2909-46e1-90ed-0063c64cef95"
// }
// }