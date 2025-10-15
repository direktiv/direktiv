import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'
import { fileURLToPath } from 'url'

import helpers from '../../common/helpers'

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test filesystem tree update paths', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCheckPathExists(it, expect, namespace, '/dir1', false)
	helpers.itShouldCheckPathExists(it, expect, namespace, '/foo1', false)

	helpers.itShouldCreateDir(it, expect, namespace, '/', 'dir1')
	helpers.itShouldCreateFile(it, expect, namespace, '/', 'foo1', 'workflow', 'text',
		btoa(helpers.dummyWorkflow('foo1')))

	helpers.itShouldCheckPathExists(it, expect, namespace, '/dir1', true)
	helpers.itShouldCheckPathExists(it, expect, namespace, '/foo1', true)

	helpers.itShouldUpdateFilePath(it, expect, namespace, '/foo1', '/foo2')

	helpers.itShouldCheckPathExists(it, expect, namespace, '/dir1', true)
	helpers.itShouldCheckPathExists(it, expect, namespace, '/foo1', false)
	helpers.itShouldCheckPathExists(it, expect, namespace, '/foo2', true)
})

describe('Test filesystem tree change dir', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateDir(it, expect, namespace, '/', 'dir1')
	helpers.itShouldCreateDir(it, expect, namespace, '/dir1', 'dir2')

	helpers.itShouldCheckPathExists(it, expect, namespace, '/dir1', true)
	helpers.itShouldCheckPathExists(it, expect, namespace, '/dir2', false)
	helpers.itShouldCheckPathExists(it, expect, namespace, '/dir1/dir2', true)

	helpers.itShouldUpdateFilePath(it, expect, namespace, '/dir1/dir2', '/dir2')

	helpers.itShouldCheckPathExists(it, expect, namespace, '/dir1', true)
	helpers.itShouldCheckPathExists(it, expect, namespace, '/dir2', true)
	helpers.itShouldCheckPathExists(it, expect, namespace, '/dir1/dir2', false)
})
