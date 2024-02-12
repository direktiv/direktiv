import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import helpers from '../common/helpers'

const namespace = basename(__filename)

describe('Test filesystem tree update paths', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCheckPathExistsV2(it, expect, namespace, '/dir1', false)
	helpers.itShouldCheckPathExistsV2(it, expect, namespace, '/foo1', false)

	helpers.itShouldCreateDirV2(it, expect, namespace, '/', 'dir1')
	helpers.itShouldCreateFileV2(it, expect, namespace, '/', 'foo1', 'workflow', 'text',
		btoa(helpers.dummyWorkflow('foo1')))

	helpers.itShouldCheckPathExistsV2(it, expect, namespace, '/dir1', true)
	helpers.itShouldCheckPathExistsV2(it, expect, namespace, '/foo1', true)


	helpers.itShouldUpdatePathV2(it, expect, namespace, '/foo1', '/foo2')

	helpers.itShouldCheckPathExistsV2(it, expect, namespace, '/dir1', true)
	helpers.itShouldCheckPathExistsV2(it, expect, namespace, '/foo1', false)
	helpers.itShouldCheckPathExistsV2(it, expect, namespace, '/foo2', true)
})


describe('Test filesystem tree change dir', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateDirV2(it, expect, namespace, '/', 'dir1')
	helpers.itShouldCreateDirV2(it, expect, namespace, '/dir1', 'dir2')

	helpers.itShouldCheckPathExistsV2(it, expect, namespace, '/dir1', true)
	helpers.itShouldCheckPathExistsV2(it, expect, namespace, '/dir2', false)
	helpers.itShouldCheckPathExistsV2(it, expect, namespace, '/dir1/dir2', true)


	helpers.itShouldUpdatePathV2(it, expect, namespace, '/dir1/dir2', '/dir2')

	helpers.itShouldCheckPathExistsV2(it, expect, namespace, '/dir1', true)
	helpers.itShouldCheckPathExistsV2(it, expect, namespace, '/dir2', true)
	helpers.itShouldCheckPathExistsV2(it, expect, namespace, '/dir1/dir2', false)
})
