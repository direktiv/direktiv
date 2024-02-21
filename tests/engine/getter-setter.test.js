import common from '../common'
import request from '../common/request'

const namespaceName = 'gettersettertest'


describe('Test getter & setter state behaviour', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	it(`should create a namespace`, async () => {
		const req = await request(common.config.getDirektivHost()).put(`/api/namespaces/${ namespaceName }`)

		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			namespace: {
				createdAt: expect.stringMatching(common.regex.timestampRegex),
				updatedAt: expect.stringMatching(common.regex.timestampRegex),
				name: namespaceName,
			},
		})
	})

	it(`should create a workflow called /test.yaml`, async () => {

		const res = await request(common.config.getDirektivHost())
			.put(`/api/namespaces/${ namespaceName }/tree/test.yaml?op=create-workflow`)
			.set({
				'Content-Type': 'text/plain',
			})
			.send(`
states:
- id: a
  type: getter
  variables:
  - key: x
    scope: namespace
    as: nsx
  - key: x
    scope: workflow
    as: wfx
  - key: x
    scope: instance
    as: inx
  transform:
    nsx: 'jq(.var.nsx // 0)'
    wfx: 'jq(.var.wfx // 0)'
    inx: 'jq(.var.inx // 0)'
  transition: b
- id: b
  type: noop
  transform: 
    nsx: 'jq(.nsx + 1)'
    wfx: 'jq(.wfx + 10)'
    inx: 'jq(.inx + 100)'
  transition: c
- id: c
  type: setter
  variables:
  - key: x
    scope: namespace
    value: 'jq(.nsx)'
  - key: x
    scope: workflow
    value: 'jq(.wfx)'
  - key: x
    scope: instance
    value: 'jq(.inx)'
`)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: namespaceName,
		})
	})

	it(`should invoke the '/test.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/test.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			nsx: 1,
			wfx: 10,
			inx: 100,
		})
	})

	it(`should invoke the '/test.yaml' workflow again`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/test.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			nsx: 2,
			wfx: 20,
			inx: 100,
		})
	})

	it(`should create a duplicate called /test2.yaml`, async () => {

		const res = await request(common.config.getDirektivHost())
			.put(`/api/namespaces/${ namespaceName }/tree/test2.yaml?op=create-workflow`)
			.set({
				'Content-Type': 'text/plain',
			})
			.send(`
states:
- id: a
  type: getter
  variables:
  - key: x
    scope: namespace
    as: nsx
  - key: x
    scope: workflow
    as: wfx
  - key: x
    scope: instance
    as: inx
  transform:
    nsx: 'jq(.var.nsx // 0)'
    wfx: 'jq(.var.wfx // 0)'
    inx: 'jq(.var.inx // 0)'
  transition: b
- id: b
  type: noop
  transform: 
    nsx: 'jq(.nsx + 1)'
    wfx: 'jq(.wfx + 10)'
    inx: 'jq(.inx + 100)'
  transition: c
- id: c
  type: setter
  variables:
  - key: x
    scope: namespace
    value: 'jq(.nsx)'
  - key: x
    scope: workflow
    value: 'jq(.wfx)'
  - key: x
    scope: instance
    value: 'jq(.inx)'
`)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: namespaceName,
		})
	})

	it(`should invoke the '/test2.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/test2.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			nsx: 3,
			wfx: 10,
			inx: 100,
		})
	})

	it(`should create a workflow called /nuller.yaml`, async () => {

		const res = await request(common.config.getDirektivHost())
			.put(`/api/namespaces/${ namespaceName }/tree/nuller.yaml?op=create-workflow`)
			.set({
				'Content-Type': 'text/plain',
			})
			.send(`
states:
- id: a
  type: setter
  variables:
  - key: x
    scope: namespace
    value: null
`)

		expect(res.statusCode).toEqual(200)
		expect(res.body).toMatchObject({
			namespace: namespaceName,
		})
	})

	it(`should invoke the '/nuller.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/nuller.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({})
	})

	it(`should invoke the '/test.yaml' workflow again`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/test.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			nsx: 1,
			wfx: 30,
			inx: 100,
		})
	})

})
