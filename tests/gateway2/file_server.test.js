import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const namespace = basename(__filename)

describe('Test file_server plugin', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)
	helpers.itShouldCreateDir(it, expect, namespace, '/', 'foo')

	helpers.itShouldCreateFile(it, expect, namespace, '/', 'file1.text', 'file', 'text/plain', btoa(`some content 11`))
	helpers.itShouldCreateFile(it, expect, namespace, '/foo/', 'file2.text', 'file', 'text/plain', btoa(`some content 22`))

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep1.yaml', 'endpoint', `
    x-direktiv-api: endpoint/v2
    x-direktiv-config:
        path: "/ep1"
        allow_anonymous: true
        plugins:
          target:
            type: target-fileserver
            configuration:
              dir: "/"
    get:
      responses:
         "200":
           description: works
    `,
	)
	retry10(`should fetch file1.text file`, async () => {
		const res = await request(config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/gateway/ep1/file1.text`)
		expect(res.statusCode).toEqual(200)
		expect(res.text).toEqual('some content 11')
		expect(res.headers['content-type']).toEqual('text/plain')
		expect(res.headers['content-length']).toEqual('15')
	})
	it(`should fetch /foo/file2.text file`, async () => {
		const res = await request(config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/gateway/ep1/foo/file2.text`)
		expect(res.statusCode).toEqual(200)
		expect(res.text).toEqual('some content 22')
		expect(res.headers['content-type']).toEqual('text/plain')
		expect(res.headers['content-length']).toEqual('15')
	})
})
