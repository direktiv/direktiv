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

describe('Runtime variables from workflow setVariable()', () => {
  beforeAll(helpers.deleteAllNamespaces)
  helpers.itShouldCreateNamespace(it, expect, namespace)

  const workflowName = 'setVariable.wf.ts'

  // Workflow code: creates a NAMESPACE-scoped runtime variable "myVar"
  const workflowSource = `
    declare function setVariable(
      scope: 'namespace' | 'workflow' | 'instance',
      name: string,
      content: string,
    ): void;

    declare function getVariable(
      scope: 'namespace' | 'workflow' | 'instance',
      name: string,
    ): string | null;

    function stateOne(payload: any) {
      const value = 'hello from runtime variable';

      // precomputed base64 of the value above so we don't need btoa() in Sobek
      const content = 'aGVsbG8gZnJvbSBydW50aW1lIHZhcmlhYmxl';

      // 1) namespace-scoped
      setVariable("namespace", "nsVar", content);
      const nsRead = getVariable("namespace", "nsVar");

      // 2) workflow-scoped
      setVariable("workflow", "wfVar", content);
      const wfRead = getVariable("workflow", "wfVar");

      // 3) instance-scoped
      setVariable("instance", "instVar", content);
      const instRead = getVariable("instance", "instVar");

      // 4) not-found case
      const missing = getVariable("namespace", "does-not-exist");

      return finish({
        ok: true,
        stored: value,
        nsRead,
        wfRead,
        instRead,
        missing,
      });
    }
  `

  // Create the TS workflow file in Direktiv
  helpers.itShouldTSWorkflow(
    it,
    expect,
    namespace,
    '/',
    workflowName,
    workflowSource,
  )

  it('should create a namespace-scoped runtime variable "myVar"', async () => {
    const baseUrl = common.config.getDirektivBaseUrl()

    // 1) Run the workflow and wait for completion
    const startRes = await request(baseUrl)
      .post(
        `/api/v2/namespaces/${namespace}/instances?` +
          `path=/${workflowName}&wait=true&fullOutput=true`,
      )
      .send({})

    expect(startRes.statusCode).toEqual(200)
    expect(startRes.body.data.status).toEqual('complete')

    // Validate getVariable outputs from the workflow result
    const output = JSON.parse(startRes.body.data.output)
    expect(output.nsRead).toEqual(encodedContent)
    expect(output.wfRead).toEqual(encodedContent)
    expect(output.instRead).toEqual(encodedContent)
    expect(output.missing).toBeNull()

    // 2) Query runtime variables for this namespace + name
    const varsRes = await request(baseUrl).get(
      `/api/v2/namespaces/${namespace}/variables?name=myVar`,
    )

    expect(varsRes.statusCode).toEqual(200)
    expect(Array.isArray(varsRes.body.data)).toBe(true)
    expect(varsRes.body.data.length).toBe(1)

    const v = varsRes.body.data[0]
    expect(v.name).toBe('myVar')
    expect(v.type).toBe('namespace-variable')
    expect(v.reference).toBe(namespace)
  })
})