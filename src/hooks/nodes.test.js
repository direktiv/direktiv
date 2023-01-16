import { renderHook, act } from "@testing-library/react-hooks";
import * as matchers from 'jest-extended';
import {useNodes} from './index'
import { Config } from "./util";
expect.extend(matchers);

// mock timer using jest
jest.useFakeTimers();

describe('useNodes', () => {
    it('fetch node',  async () => {
        const { result, waitForNextUpdate } = renderHook(() => useNodes(Config.url, false, Config.namespace, ""));
        await waitForNextUpdate()
        expect(result.current.data.node.path).toEqual("/")
    })
    it('stream node', async() => {
        const { result, waitForNextUpdate } = renderHook(() => useNodes(Config.url, true, Config.namespace, ""));
        await waitForNextUpdate()
        expect(result.current.data.node.path).toEqual("/")
    })
    it('fetch node that doesnt exist', async () => {
        const { result, waitForNextUpdate } = renderHook(() => useNodes(Config.url, false, Config.namespace, "testtest"));
        await waitForNextUpdate()
        expect(result.current.err).not.toBeNull()
    })
    it('create directory, rename directory then delete directory', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useNodes(Config.url, true, Config.namespace, ""));
        await waitForNextUpdate()
        
        act(()=>{
             result.current.createNode("/test-directory", "directory")
        })
        await waitForNextUpdate()
 
        let found = false
        for (var i=0; i < result.current.data.children.edges.length; i++) {
            if(result.current.data.children.edges[i].node.name === "test-directory"){
                found = true
            }
        }

        expect(found).toBeTrue()

        act(()=> {
            result.current.renameNode("", "/test-directory", "/test-directory2")
        })

        await waitForNextUpdate()
        
        found = false
        for (var i=0; i < result.current.data.children.edges.length; i++) {
            if(result.current.data.children.edges[i].node.name === "test-directory2"){
                found = true
            }
        }

        expect(found).toBeTrue()
    })
    it('create directory check if exist then delete', async ()=> {
        const { result, waitForNextUpdate } = renderHook(() => useNodes(Config.url, true, Config.namespace, ""));
        await waitForNextUpdate()
        
        await act( async()=>{
            await result.current.createNode("/test-directory", "directory")
        })
        await waitForNextUpdate()
        let found = false
        for (var i=0; i < result.current.data.children.edges.length; i++) {
            if(result.current.data.children.edges[i].node.name === "test-directory"){
                found = true
            }
        }

        expect(found).toBeTrue()

        await act( async()=> {
            await result.current.deleteNode("test-directory")
        })

        await waitForNextUpdate()

        found = false
        for (var i=0; i < result.current.data.children.edges.length; i++) {
            if(result.current.data.children.edges[i].node.name === "test-directory"){
                found = true
            }
        }
        expect(found).not.toBeTrue()
    })
    it('create workflow then toggle it to false and check if its false', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useNodes(Config.url, true, Config.namespace, ""));
        
        await waitForNextUpdate()
        
        await act( async ()=>{
            await result.current.createNode("/test-workflow", "workflow",  `states:
            - id: helloworld
              type: noop
              transform:
                result: Hello world!
            `)
        })

        await waitForNextUpdate()

        let found = false
        for (var i=0; i < result.current.data.children.edges.length; i++) {
            if(result.current.data.children.edges[i].node.name === "test-workflow"){
                found = true
            }
        }

        expect(found).toBeTrue()

        // toggle the workflow here
        await act (async()=>{
            await result.current.toggleWorkflow("test-workflow", false)
        })

        // get workflow router
        expect(await result.current.getWorkflowRouter("test-workflow")).not.toBeTrue()

        act( ()=> {
            result.current.deleteNode("test-workflow")
        })

        await waitForNextUpdate()

        found = false
        for (var i=0; i < result.current.data.children.edges.length; i++) {
            if(result.current.data.children.edges[i].node.name === "test-workflow"){
                found = true
            }
        }
        expect(found).not.toBeTrue()


 
    })
    it('create workflow check if exist then delete', async () => {
        const { result, waitForNextUpdate } = renderHook(() => useNodes(Config.url, true, Config.namespace, ""));
        
        await waitForNextUpdate()
        
        await act( async()=>{
            await result.current.createNode("/test-workflow", "workflow",  `states:
            - id: helloworld
              type: noop
              transform:
                result: Hello world!
            `)
        })

        await waitForNextUpdate()
 
        let found = false
        for (var i=0; i < result.current.data.children.edges.length; i++) {
            if(result.current.data.children.edges[i].node.name === "test-workflow"){
                found = true
            }
        }

        expect(found).toBeTrue()
        await act( async()=> {
            await result.current.deleteNode("test-workflow")
        })

        await waitForNextUpdate()

        found = false
        for (var i=0; i < result.current.data.children.edges.length; i++) {
            if(result.current.data.children.edges[i].node.name === "test-workflow"){
                found = true
            }
        }
        expect(found).not.toBeTrue()
    })
})