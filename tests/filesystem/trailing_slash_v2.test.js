import { beforeAll, describe, expect, it } from '@jest/globals'

import helpers from '../common/helpers'
import {basename} from "path";

const namespace = basename(__filename)

describe('Test filesystem tree read operations', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)
	helpers.itShouldCreateFileV2(it, expect, namespace,
		'',
		'foo1',
		'workflow',
		'text',
		btoa(helpers.dummyWorkflow('foo1')))

	helpers.itShouldCreateFileV2(it, expect, namespace,
		'/',
		'foo2',
		'workflow',
		'text',
		btoa(helpers.dummyWorkflow('foo2')))

	helpers.itShouldCreateDirV2(it, expect, namespace, '', 'dir1')
	helpers.itShouldCreateDirV2(it, expect, namespace, '/', 'dir2')
})
