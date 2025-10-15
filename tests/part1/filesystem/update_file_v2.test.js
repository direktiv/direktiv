import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import { fileURLToPath } from 'url'

import helpers from '../../common/helpers'

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test filesystem tree update operations', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateDir(it, expect, namespace, '/', 'dir1')
	helpers.itShouldCreateFile(it, expect, namespace,
		'/dir1',
		'foo1',
		'workflow',
		'text/plain',
		btoa(helpers.dummyWorkflow('foo1')))

	helpers.itShouldUpdateFile(it, expect, namespace,
		'/dir1/foo1',
		{
			path: '/dir1/foo2',
			data: btoa(helpers.dummyWorkflow('foo2')),
		},
	)
})
