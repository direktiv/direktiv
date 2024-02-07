import { describe, expect, it } from '@jest/globals'
import helpers from '../common/helpers'

const testNamespace = 'test-file-namespace'

describe('Test filesystem tree update operations', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, testNamespace)

	helpers.itShouldCreateDirV2(it, expect, testNamespace, '/', 'dir1')
	helpers.itShouldCreateFileV2(it, expect, testNamespace,
		'/dir1',
		'foo1',
		'workflow',
		'text/plain',
		btoa(helpers.dummyWorkflow('foo1')))

	helpers.itShouldUpdateFileV2(it, expect, testNamespace,
		'/dir1/foo1',
		{ absolutePath: '/dir1/foo2',
			data: btoa(helpers.dummyWorkflow('foo2')) },
	)
})
