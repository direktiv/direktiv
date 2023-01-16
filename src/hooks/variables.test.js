import { renderHook, act } from "@testing-library/react-hooks";
import * as matchers from 'jest-extended';
import {useNamespaceVariables, useNodes, useWorkflowVariables} from './index'
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

beforeAll(async ()=>{
    console.log('creating workflow')
    const { result, waitForNextUpdate} = renderHook(()=> useNodes(Config.url, true, Config.namespace, ""))
    await waitForNextUpdate()
    act( ()=>{
        result.current.createNode("/test-workflowvar-hook", "workflow",  wfyaml)
    })
    await waitForNextUpdate()
})

describe('useNamespaceVariables', ()=>{
    it('fetch namespace variables', async ()=>{
        const { result, waitForNextUpdate } = renderHook(() => useNamespaceVariables(Config.url, false, Config.namespace));
        await waitForNextUpdate()
        expect(result.current.data.length).toEqual(0)
    })
    it('stream namespace variables', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useNamespaceVariables(Config.url, true, Config.namespace))
        await waitForNextUpdate()
        expect(result.current.data.length).toEqual(0)
    })
    it('set namespace variable, get namespace variable and then delete namespace variable', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useNamespaceVariables(Config.url, true, Config.namespace))
        await waitForNextUpdate()
        expect(result.current.data.length).toEqual(0)
        
        act(()=>{
            result.current.setNamespaceVariable("testnamespace", "testnamespace")
        })

        await waitForNextUpdate()
        let found = false
        for(var i=0; i < result.current.data.length; i++) {
            if(result.current.data[i].node.name === "testnamespace"){
                found = true
            }
        }
        expect(found).toBeTrue()

        // get workflow variable data
        let testvar = await result.current.getNamespaceVariable("testnamespace")
        expect(testvar.data).toEqual("testnamespace")
        expect(testvar.contentType).toEqual("application/json")

        // delete workflow variable
        act(()=>{
            result.current.deleteNamespaceVariable("testnamespace")
        })

        await waitForNextUpdate()
        found = false
        for(var i=0; i < result.current.data.length; i++) {
            if(result.current.data[i].node.name === "testnamespace"){
                found = true
            }
        }
        expect(found).not.toBeTrue()

    })
})

describe('useWorkflowVariables', ()=>{
    it('fetch workflow variables', async ()=>{
        const { result, waitForNextUpdate } = renderHook(() => useWorkflowVariables(Config.url, false, Config.namespace, "test-workflowvar-hook"));
        await waitForNextUpdate()
        expect(result.current.data.length).toEqual(0)
    })
    it('stream workflow variables', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useWorkflowVariables(Config.url, true, Config.namespace, "test-workflowvar-hook"));
        await waitForNextUpdate()
        expect(result.current.data.length).toEqual(0)
    })
    it('set workflow variable, get workflow variable data and then delete workflow variable', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useWorkflowVariables(Config.url, true, Config.namespace, "test-workflowvar-hook"));
        await waitForNextUpdate()
        expect(result.current.data.length).toEqual(0)

        // set workflow variable
        act(()=>{
            result.current.setWorkflowVariable("test", "test")
        })

        await waitForNextUpdate()
        let found = false
        for(var i=0; i < result.current.data.length; i++) {
            if(result.current.data[i].node.name === "test"){
                found = true
            }
        }
        expect(found).toBeTrue()

        // get workflow variable data
        let testvar = await result.current.getWorkflowVariable("test")
        expect(testvar.data).toEqual("test")
        expect(testvar.contentType).toEqual("application/json")

        // delete workflow variable
        act(()=>{
            result.current.deleteWorkflowVariable("test")
        })

        await waitForNextUpdate()
        found = false
        for(var i=0; i < result.current.data.length; i++) {
            if(result.current.data[i].node.name === "test"){
                found = true
            }
        }
        expect(found).not.toBeTrue()
    })
})

afterAll(async ()=>{
    console.log('deleting workflow')

    const { result, waitForNextUpdate} = renderHook(()=> useNodes(Config.url, true, Config.namespace, ""))
    await waitForNextUpdate()
    act( ()=>{
        result.current.deleteNode('test-workflowvar-hook')
    })
    await waitForNextUpdate()
})