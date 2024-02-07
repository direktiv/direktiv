import { beforeAll, describe, expect, it } from '@jest/globals'

import helpers from '../common/helpers'

const testNamespace = 'test-file-namespace'

describe('Test filesystem tree read operations', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, testNamespace)
	helpers.itShouldCreateFileV2(it, expect, testNamespace,
		'',
		'foo1',
		'workflow',
		'text',
		btoa(helpers.dummyWorkflow('foo1')))

	helpers.itShouldCreateFileV2(it, expect, testNamespace,
		'/',
		'foo2',
		'workflow',
		'text',
		btoa(helpers.dummyWorkflow('foo2')))

	helpers.itShouldCreateDirV2(it, expect, testNamespace, '', 'dir1')
	helpers.itShouldCreateDirV2(it, expect, testNamespace, '/', 'dir2')
})
