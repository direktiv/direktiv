import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespaceName = 'gettersettertest'

describe('Test getter & setter state behaviour', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'test.yaml',
		'workflow',
		'text/plain',
		btoa(`
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
`))

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

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'test2.yaml',
		'workflow',
		'text/plain',
		btoa(`
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
`))

	it(`should invoke the '/test2.yaml' workflow`, async () => {
		const req = await request(common.config.getDirektivHost()).get(`/api/namespaces/${ namespaceName }/tree/test2.yaml?op=wait`)
		expect(req.statusCode).toEqual(200)
		expect(req.body).toMatchObject({
			nsx: 3,
			wfx: 10,
			inx: 100,
		})
	})

	helpers.itShouldCreateFileV2(it, expect, namespaceName,
		'',
		'nuller.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: setter
  variables:
  - key: x
    scope: namespace
    value: null
`))

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
