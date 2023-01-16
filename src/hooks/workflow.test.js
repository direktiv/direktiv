import { renderHook, act } from "@testing-library/react-hooks";
import * as matchers from 'jest-extended';
import {useNodes, useWorkflow, useWorkflowLogs} from './index'
import { Config } from "./util";
expect.extend(matchers);

// mock timer using jest
jest.useFakeTimers();

const wfyaml = `states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!
`

const wfyaml2 = `states:
- id: helloworld2
  type: noop
  transform:
    result: Hello world2!
`

const wfyaml3 = `states:
- id: helloworld2revs
  type: noop
  transform:
    result: Hello world2!
`


beforeAll(async ()=>{
    console.log('creating workflow')
    const { result, waitForNextUpdate} = renderHook(()=> useNodes(Config.url, true, Config.namespace, ""))
    await waitForNextUpdate()
    await act( async()=>{
        await result.current.createNode("/test-workflow-hook", "workflow",  wfyaml)
    })
})

describe('useWorkflow', () => {
    it('fetch workflow',  async () => {
        const { result, waitForNextUpdate } = renderHook(() => useWorkflow(Config.url, false, Config.namespace, "test-workflow-hook"));
        await waitForNextUpdate()
        expect(result.current.data.revision.source).not.toEqual("")
    })
    it('stream  workflow', async() => {
        const { result, waitForNextUpdate } = renderHook(() => useWorkflow(Config.url, true, Config.namespace, "test-workflow-hook"));
        await waitForNextUpdate()
        expect(result.current.data.revision.source).not.toEqual("")
    })
    it('update workflow', async() => {
        const { result, waitForNextUpdate } = renderHook(() => useWorkflow(Config.url, true, Config.namespace, "test-workflow-hook"));
        await waitForNextUpdate()
        act(()=>{
            result.current.updateWorkflow(wfyaml2)
        })
        await waitForNextUpdate()
        expect(atob(result.current.data.revision.source)).toEqual(wfyaml2)
    })
    it('add workflow attributes', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useWorkflow(Config.url, false, Config.namespace, "test-workflow-hook"));
        await waitForNextUpdate()
        await act(async()=>{
            await result.current.addAttributes(["test", "test2"])
            await result.current.getWorkflow()
        })
        expect(result.current.data.node.attributes).toEqual(["test", "test2"])

        await act(async()=>{
            await result.current.deleteAttributes(["test", "test2"])
            await result.current.getWorkflow()
        })

        expect(result.current.data.node.attributes.length).toEqual(0)
    })
    it('execute workflow', async ()=>{
        const { result, waitForNextUpdate } = renderHook(() => useWorkflow(Config.url, true, Config.namespace, "test-workflow-hook"));
        await waitForNextUpdate()
        let instanceId = await result.current.executeWorkflow()
        expect(instanceId).not.toEqual("")
    })
    it('list instances for workflow', async ()=> {
        const { result, waitForNextUpdate } = renderHook(() => useWorkflow(Config.url, true, Config.namespace, "test-workflow-hook"));
        await waitForNextUpdate()
        let instances = await result.current.getInstancesForWorkflow()
        expect(instances[0].node.as).toEqual("test-workflow-hook")
    })
    it('toggle workflow and check router to see active is false', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useWorkflow(Config.url, true, Config.namespace, "test-workflow-hook"));
        await waitForNextUpdate()
        await act (async()=>{
            await result.current.toggleWorkflow(false)
        })
        let json = await result.current.getWorkflowRouter()
        expect(json.live).not.toBeTrue()
        await act (async()=>{
            await result.current.toggleWorkflow(true)
        })
        json = await result.current.getWorkflowRouter()
        expect(json.live).toBeTrue()
    })
    it('edit workflow router set active to true', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useWorkflow(Config.url, true, Config.namespace, "test-workflow-hook"));
        await waitForNextUpdate()
        await act (async()=>{
            await result.current.editWorkflowRouter([], true)
        })
        let json = await result.current.getWorkflowRouter()
        expect(json.live).toBeTrue()
    })
    it('set workflow to log to event', async ()=> {
        const { result, waitForNextUpdate } = renderHook(() => useWorkflow(Config.url, false, Config.namespace, "test-workflow-hook"));
        await waitForNextUpdate()
        await act (async ()=>{
           await result.current.setWorkflowLogToEvent("test")
           await result.current.getWorkflow()
        })
        expect(result.current.data.eventLogging).toEqual("test")
    })
    it('get state metrics for workflow', async ()=>{
        const { result, waitForNextUpdate } = renderHook(() => useWorkflow(Config.url, false, Config.namespace, "test-workflow-hook"));
        await waitForNextUpdate()
        expect(await result.current.getStateMillisecondMetrics()).toBeInstanceOf(Array)
    })
    it('fetch workflow logs', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useWorkflowLogs(Config.url, false, Config.namespace, "test-workflow-hook"));
        await waitForNextUpdate()
        expect(result.current.data).toBeInstanceOf(Array)
    })
    it('stream workflow logs', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useWorkflowLogs(Config.url, true, Config.namespace, "test-workflow-hook"));
        await waitForNextUpdate()
        expect(result.current.data).toBeInstanceOf(Array)
    })
    it('get revisions for workflow, tag latest as a version', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useWorkflow(Config.url, true, Config.namespace, "test-workflow-hook"));
        await waitForNextUpdate()
        let revisions = await result.current.getRevisions()
        expect(revisions[0].node.name).toBe("latest")

        await act(async()=>{
            await result.current.tagWorkflow("latest", "latest2")
        })
        revisions = await result.current.getRevisions()
        let found = false
        for(var i=0; i < revisions.length; i++) {
            if(revisions[i].node.name === "latest2") {
                found = true
            }
        }
        expect(found).toBeTrue()
    })
    it('get revisions, update workflow, delete revision, update back to old flow then discard', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useWorkflow(Config.url, true, Config.namespace, "test-workflow-hook"));
        await waitForNextUpdate()
        let revisions = await result.current.getRevisions()
        expect(revisions[0].node.name).toBe("latest")

        let rev = null
        await act(async()=>{
            await result.current.updateWorkflow(wfyaml3)
            rev = await result.current.saveWorkflow()
            revisions = await result.current.getRevisions()
        })

        let revlength = revisions.length

        for(var i=0; i < revisions.length; i++) {
            if(revisions[i].node.name !== "latest2" && revisions[i].node.name !== "latest") {
                await result.current.deleteRevision(revisions[i].node.name)
            }
        }

        revisions = await result.current.getRevisions()

        expect(revisions.length).not.toBe(revlength - 1)

        await act(async()=>{
            await result.current.updateWorkflow(wfyaml)
            await result.current.discardWorkflow()
            await result.current.getWorkflow()
        })
        expect(atob(result.current.data.revision.source)).toBe(wfyaml3)
    })
})


afterAll(async ()=>{
    console.log('deleting workflow')

    const { result, waitForNextUpdate} = renderHook(()=> useNodes(Config.url, true, Config.namespace, ""))
    await waitForNextUpdate()
    act( ()=>{
        result.current.deleteNode('test-workflow-hook')
    })
    await waitForNextUpdate()
})