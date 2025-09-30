import { beforeAll, describe, expect, it } from '@jest/globals'

import common from '../common'
import helpers from '../common/helpers'

const namespaceName = 'executetest'

describe('test execute workflow', () => {
	beforeAll(common.helpers.deleteAllNamespaces)

	helpers.itShouldCreateNamespace(it, expect, namespaceName)

	helpers.itShouldCreateFile(it, expect, namespaceName,
		'',
		'flow.wf.ts',
		'workflow',
		'text/typescript',
		btoa(`
jens`))
})
