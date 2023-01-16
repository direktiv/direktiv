import { renderHook, act } from "@testing-library/react-hooks";
import * as matchers from 'jest-extended';
import {useInstance, useInstanceLogs, useNodes, useWorkflow, useWorkflowLogs} from './index'
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
const delayyml = `description: A simple 'delay' state that waits for 5 seconds
states:
- id: delay
  type: delay
  duration: PT5S`

let createWF = async () => {
    console.log('create workflow')
    const { result, waitForNextUpdate} = renderHook(()=> useNodes(Config.url, true, Config.namespace, ""))
    await waitForNextUpdate()
    await act( async()=>{
        await result.current.createNode("/test-workflow-hook-execute", "workflow",  wfyaml)
        await result.current.createNode("/test-delay", "workflow", delayyml)
    })
}

let executeWF = async() => {
    console.log('execute workflow')
    const { result, waitForNextUpdate} = renderHook(()=> useWorkflow(Config.url, true, Config.namespace, "test-workflow-hook-execute"))
    await waitForNextUpdate()
    await act( async()=>{
        process.env.INSTANCE_ID = await result.current.executeWorkflow(JSON.stringify({"test":"hello"}))
    })
}
let executeDelayWF = async()=>{
    console.log('execute delay workflow')
    const { result, waitForNextUpdate} = renderHook(()=> useWorkflow(Config.url, true, Config.namespace, "test-delay"))
    await waitForNextUpdate()
    await act( async()=>{
        process.env.DELAY_INSTANCE_ID = await result.current.executeWorkflow()
    })
}

beforeAll(async ()=>{
    await createWF()
    await executeWF()
    await executeDelayWF()
})

describe('useInstance', ()=>{
    it('stream instance details',async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useInstance(Config.url, true, Config.namespace, process.env.INSTANCE_ID))
        await waitForNextUpdate()
        expect(result.current.data.as).toEqual("test-workflow-hook-execute")
    })
    it('fetch instance details',async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useInstance(Config.url, false, Config.namespace, process.env.INSTANCE_ID))
        await waitForNextUpdate()
        expect(result.current.data.as).toEqual("test-workflow-hook-execute")
    })
    it('get input', async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useInstance(Config.url, false, Config.namespace, process.env.INSTANCE_ID))
        await waitForNextUpdate()
        let input = await result.current.getInput()
        expect(JSON.parse(input).test).toBe("hello")
    })
    it('get output', async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useInstance(Config.url, false, Config.namespace, process.env.INSTANCE_ID))
        await waitForNextUpdate()
        let output = await result.current.getOutput()
        expect(JSON.parse(output).result).toBe("Hello world!")
    })
    it('execute delay wf and cancel check status to make sure it was cancelled', async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useInstance(Config.url, true, Config.namespace, process.env.DELAY_INSTANCE_ID))
        await waitForNextUpdate()
        console.log(process.env.DELAY_INSTANCE_ID)
        act(()=>{
            result.current.cancelInstance()
        })
        await waitForNextUpdate()
        expect(result.current.data.errorCode).toBe("direktiv.cancels.api")
    })
})

describe('useInstanceLogs', ()=>{
    it('stream instance logs', async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useInstanceLogs(Config.url, true, Config.namespace, process.env.INSTANCE_ID))
        await waitForNextUpdate()
        expect(result.current.data).toBeInstanceOf(Array)
    })
    it('fetch instance logs', async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useInstanceLogs(Config.url, true, Config.namespace, process.env.INSTANCE_ID))
        await waitForNextUpdate()
        expect(result.current.data).toBeInstanceOf(Array)
    })
})

afterAll(async ()=>{
    console.log('deleting workflow')
    const { result, waitForNextUpdate} = renderHook(()=> useNodes(Config.url, true, Config.namespace, ""))
    await waitForNextUpdate()
    act( ()=>{
        result.current.deleteNode('test-workflow-hook-execute')
        result.current.deleteNode('test-delay')
    })
    await waitForNextUpdate()
})

