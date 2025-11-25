import { beforeAll, describe, expect, it } from '@jest/globals'
import { btoa } from 'js-base64'
import { basename } from 'path'
import { fileURLToPath } from 'url'

import common from '../common'
import helpers from '../common/helpers'
import request from '../common/request'

const namespace =
	helpers.randomLowercaseString(3) +
	'-' +
	basename(fileURLToPath(import.meta.url))

describe('Test cron workflows', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	const files = [
		{
			name: 'basic1Sec.wf.ts',
			file: `
		const flow: FlowDefinition = {
  			cron: "*/1 * * * * *"
		};
		function stateOne(payload) {
			payload.foo = "1sec";
			return finish(payload);
		}
`,
		},
		{
			name: 'basic2Sec.wf.ts',
			file: `
		const flow: FlowDefinition = {
  			cron: "*/2 * * * * *"
		};
		function stateOne(payload) {
			payload.foo = "2sec";
			return finish(payload);
		}
`,
		},
	]

	for (let i = 0; i < files.length; i++) {
		const file = files[i]
		helpers.itShouldCreateFile(
			it,
			expect,
			namespace,
			'/',
			file.name,
			'workflow',
			'application/x-typescript',
			btoa(file.file),
		)
	}

	it(`should list instances`, async () => {
		await helpers.sleep(7000)
		const res = await request(common.config.getDirektivBaseUrl()).get(
			`/api/v2/namespaces/${namespace}/instances`,
		)
		expect(res.statusCode).toEqual(200)
		expect(res.body.data.every((item) => item.status === 'complete')).toEqual(
			true,
		)
		expect(res.body.data.every((item) => item.invoker === 'cron')).toEqual(true)
		expect(res.body.data.every((item) => item.outputLength === 14)).toEqual(
			true,
		)

		let goodLength = res.body.data.length >= 10 && res.body.data.length <= 14
		expect(goodLength).toEqual(true)

		let mapped = res.body.data.map((item) => item.path + item.output)

		let sec1 = 0
		let sec2 = 0
		for (let i = 0; i < mapped.length; i++) {
			if (mapped[i] === '/basic1Sec.wf.ts{"foo":"1sec"}') {
				sec1++
			}
			if (mapped[i] === '/basic2Sec.wf.ts{"foo":"2sec"}') {
				sec2++
			}
		}
		expect(sec1 + sec2).toEqual(res.body.data.length)
		expect(sec2 / sec1).toBeGreaterThan(0.4)
		expect(sec2 / sec1).toBeLessThan(0.6)
	})
})
