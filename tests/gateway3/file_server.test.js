import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'
import {fileURLToPath} from "url";

const namespace = basename(fileURLToPath(import.meta.url))

describe('Test file_server plugin', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)
	helpers.itShouldCreateDir(it, expect, namespace, '/', 'foo')
	helpers.itShouldCreateDir(it, expect, namespace, '/', 'bar')

	helpers.itShouldCreateFile(it, expect, namespace, '/', 'file1.text', 'file', 'text/plain', btoa(`some content 11`))
	helpers.itShouldCreateFile(it, expect, namespace, '/foo/', 'file2.text', 'file', 'text/plain', btoa(`some content 22`))
	helpers.itShouldCreateFile(it, expect, namespace, '/foo/', 'file3.text', 'file', 'text/plain', btoa(`some content 33`))
	helpers.itShouldCreateFile(it, expect, namespace, '/bar/', 'file4.text', 'file', 'text/plain', btoa(`some content 44`))

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
              allow_paths: ["/foo", "/file1.text"]
              deny_paths: ["/bar"]
    get:
      responses:
         "200":
           description: works
    `,
	)
	retry10(`should fetch file1.text file`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(`/api/v2/namespaces/${ namespace }/gateway/ep1/file1.text`)
		expect(res.statusCode).toEqual(200)
		expect(res.text).toEqual('some content 11')
		expect(res.headers['content-type']).toEqual('text/plain')
		expect(res.headers['content-length']).toEqual('15')
	})
	it(`should fetch /foo/file2.text file`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(`/api/v2/namespaces/${ namespace }/gateway/ep1/foo/file2.text`)
		expect(res.statusCode).toEqual(200)
		expect(res.text).toEqual('some content 22')
		expect(res.headers['content-type']).toEqual('text/plain')
		expect(res.headers['content-length']).toEqual('15')
	})
	it(`should deny /bar/file4.text file`, async () => {
		const res = await request(config.getDirektivBaseUrl()).get(`/api/v2/namespaces/${ namespace }/gateway/ep1/bar/file4.text`)
		expect(res.statusCode).toEqual(500)
		expect(res.text).toEqual('{"error":{"endpointFile":"/ep1.yaml","message":"path is denied"}}\n')
	})
})
