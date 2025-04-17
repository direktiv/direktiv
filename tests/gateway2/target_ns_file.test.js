import { beforeAll, describe, expect, it } from '@jest/globals'
import { basename } from 'path'

import config from '../common/config'
import helpers from '../common/helpers'
import request from '../common/request'
import { retry10 } from '../common/retry'

const namespace = basename(__filename)

describe('Test target-namespace-file plugin', () => {
	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)

	helpers.itShouldCreateFile(it, expect, namespace, '/', 'some.text', 'file', 'text/plain', btoa(`some content`))

	helpers.itShouldCreateYamlFile(it, expect, namespace,
		'/', 'ep1.yaml', 'endpoint', `
    x-direktiv-api: endpoint/v2
    x-direktiv-config:
        path: /ep1
        allow_anonymous: true
        plugins:
            target:
                type: target-namespace-file
                configuration:
                    namespace: ${ namespace }
                    file: /some.text
    get:
        responses:
            "200":
            description: works
    `,
	)

	retry10(`should fetch some.text file`, async () => {
    	const res = await request(config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/gateway/ep1`)
    	expect(res.statusCode).toEqual(200)
    	expect(res.text).toEqual('some content')
    	expect(res.headers['content-type']).toEqual('text/plain')
    	expect(res.headers['content-length']).toEqual('12')
	})

	// test system namespace access.
	helpers.itShouldCreateNamespace(it, expect, 'system')

	helpers.itShouldCreateYamlFile(it, expect, 'system',
		'/', 'ep2.yaml', 'endpoint', `
    x-direktiv-api: endpoint/v2
    x-direktiv-config:
        path: /ep2
        allow_anonymous: true
        plugins:
            target:
                type: target-namespace-file
                configuration:
                    namespace: ${ namespace }
                    file: /some.text
    get:
        responses:
            "200":
            description: works
    `,
	)

	retry10(`should fetch some.text file from system namespace`, async () => {
    	const res = await request(config.getDirektivHost()).get(`/api/v2/namespaces/system/gateway/ep2`)
    	expect(res.statusCode).toEqual(200)
    	expect(res.text).toEqual('some content')
    	expect(res.headers['content-type']).toEqual('text/plain')
    	expect(res.headers['content-length']).toEqual('12')
	})

	// test access denied of different namespace
	const otherNamespace = namespace + '_different'
	helpers.itShouldCreateNamespace(it, expect, otherNamespace)

	helpers.itShouldCreateYamlFile(it, expect, otherNamespace,
		'/', 'ep3.yaml', 'endpoint', `
    x-direktiv-api: endpoint/v2
    x-direktiv-config:
        path: /ep3
        allow_anonymous: true
        plugins:
            target:
                type: target-namespace-file
                configuration:
                    namespace: ${ namespace }
                    file: /some.text
    get:
        responses:
            "200":
            description: works
    `,
	)

	retry10(`should deny access fetching some.text file from different namespace`, async () => {
    	const res = await request(config.getDirektivHost()).get(`/api/v2/namespaces/${ otherNamespace }/gateway/ep3`)
    	expect(res.statusCode).toEqual(403)
	})
})
