import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import helpers from '../../common/helpers'
import {fileURLToPath} from "url";

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test filesystem tree read operations', () => {
	beforeAll(helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespace)
	helpers.itShouldCreateFile(it, expect, namespace,
		'',
		'foo1',
		'workflow',
		'text',
		btoa(helpers.dummyWorkflow('foo1')))

	helpers.itShouldCreateFile(it, expect, namespace,
		'/',
		'foo2',
		'workflow',
		'text',
		btoa(helpers.dummyWorkflow('foo2')))

	helpers.itShouldCreateDir(it, expect, namespace, '', 'dir1')
	helpers.itShouldCreateDir(it, expect, namespace, '/', 'dir2')
})
