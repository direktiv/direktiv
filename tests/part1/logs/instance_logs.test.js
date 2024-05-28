import { beforeAll, describe, expect, it } from '@jest/globals'

import { basename } from 'path'
import common from '../../common'
import helpers from '../../common/helpers'
import request from '../../common/request'
import { retry50 } from '../../common/retry'

const namespace = basename(__filename)

describe('Test instance log api calls', () => {

	beforeAll(helpers.deleteAllNamespaces)
	helpers.itShouldCreateNamespace(it, expect, namespace)
	helpers.itShouldCreateFileV2(it, expect, namespace,
		'',
		'noop.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: noop
  log: "This Is A Test"
  transform:
    result: x`))


    
	it(`generate some logs`, async () => {
		const res = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/instances?path=noop.yaml&wait=true`)
		expect(res.statusCode).toEqual(200)
	})

	retry50(`should contain instance log entries`, async () => {

		const instRes = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/instances`)
		expect(instRes.statusCode).toEqual(200)
	
		const logRes = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/logs?instance=${ instRes.body.data[0].id }`)
		expect(logRes.statusCode).toEqual(200)

        expect(logRes.body.data).toEqual(          
         expect.arrayContaining([      
          expect.objectContaining({   
            msg: 'Workflow completed.'               
          })
        ])
        )

        expect(logRes.body.data).toEqual(          
            expect.arrayContaining([      
             expect.objectContaining({   
               msg: '"This Is A Test"'               
             })
          ])
        )
	},
	)

    helpers.itShouldCreateFileV2(it, expect, namespace,
		'',
		'noop-error.yaml',
		'workflow',
		'text/plain',
		btoa(`
states:
- id: a
  type: noop
  transform:
    result: jq(.doesnotexist)`))


    
	it(`generate some logs for error`, async () => {
		const res = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/instances?path=noop-error.yaml&wait=true`)
		expect(res.statusCode).toEqual(500)
        expect(res.headers["direktiv-instance-error-code"]).toEqual("direktiv.jq.badCommand")
	})


	retry50(`should contain instance log entries`, async () => {

        const instRes = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/instances?filter.field=AS&filter.type=CONTAINS&filter.val=noop-error`)
        expect(instRes.statusCode).toEqual(200)
	
		const logRes = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/logs?instance=${ instRes.body.data[0].id }`)
		expect(logRes.statusCode).toEqual(200)

        expect(logRes.body.data).toEqual(          
         expect.arrayContaining([      
          expect.objectContaining({   
            level: 'ERROR'               
          })
        ])
        )
	},
	)

    helpers.itShouldCreateFileV2(it, expect, namespace,
		'',
		'action-error.yaml',
		'workflow',
		'text/plain',
		btoa(`
direktiv_api: workflow/v1
functions:
- id: get
  image: direktiv/request:v4
  type: knative-workflow
states:
- id: getter 
  type: action
  action:
    function: get
    input: 
      method: "DOESNTWORK"
      url: "invalid"
`))

    it(`generate some logs for error`, async () => {
        const res = await request(common.config.getDirektivHost()).post(`/api/v2/namespaces/${ namespace }/instances?path=action-error.yaml&wait=true`)
        expect(res.statusCode).toEqual(500)
        expect(res.headers["direktiv-instance-error-code"]).toEqual("com.send-request.error")
    })


    retry50(`should contain instance log entries`, async () => {

      const instRes = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/instances?filter.field=AS&filter.type=CONTAINS&filter.val=action-error`)
      expect(instRes.statusCode).toEqual(200)
    
      const logRes = await request(common.config.getDirektivHost()).get(`/api/v2/namespaces/${ namespace }/logs?instance=${ instRes.body.data[0].id }`)
      expect(logRes.statusCode).toEqual(200)

      expect(logRes.body.data).toEqual(          
      expect.arrayContaining([      
        expect.objectContaining({   
          level: 'ERROR'               
        })
      ])
      )
    },
    )



})



