import { beforeAll, describe, expect, it } from '@jest/globals'

import helpers from '../common/helpers'

const testNamespace = 'test-file-namespace'

describe('Test filesystem tree update paths', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, testNamespace)

	helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, '/dir1', false)
	helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, '/foo1', false)

	helpers.itShouldCreateDirV2(it, expect, testNamespace, '/', 'dir1')
	helpers.itShouldCreateFileV2(it, expect, testNamespace, '/', 'foo1', 'workflow', 'text',
		btoa(helpers.dummyWorkflow('foo1')))

	helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, '/dir1', true)
	helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, '/foo1', true)


	helpers.itShouldUpdatePathV2(it, expect, testNamespace, '/foo1', '/foo2')

	helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, '/dir1', true)
	helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, '/foo1', false)
	helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, '/foo2', true)
})


describe('Test filesystem tree change dir', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, testNamespace)

	helpers.itShouldCreateDirV2(it, expect, testNamespace, '/', 'dir1')
	helpers.itShouldCreateDirV2(it, expect, testNamespace, '/dir1', 'dir2')

	helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, '/dir1', true)
	helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, '/dir2', false)
	helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, '/dir1/dir2', true)


	helpers.itShouldUpdatePathV2(it, expect, testNamespace, '/dir1/dir2', '/dir2')

	helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, '/dir1', true)
	helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, '/dir2', true)
	helpers.itShouldCheckPathExistsV2(it, expect, testNamespace, '/dir1/dir2', false)
})
