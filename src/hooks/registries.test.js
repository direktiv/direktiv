import { renderHook, act } from "@testing-library/react-hooks";
import * as matchers from 'jest-extended';
import { Config } from './util';
import { useRegistries, useGlobalPrivateRegistries, useGlobalRegistries } from './index' 
expect.extend(matchers);

// mock timer using jest
jest.useFakeTimers();

describe('useRegistries', () => {
    it('list registries',  async () => {
        const { result, waitForNextUpdate } = renderHook(() => useRegistries(Config.url, Config.namespace));
        await waitForNextUpdate()
        
        expect(result.current.data).toBeArray()
    })
    it('create and delete registry', async () => {
        const { result, waitForNextUpdate } = renderHook(() => useRegistries(Config.url, Config.namespace));
        await waitForNextUpdate()

        expect(result.current.data).toBeArray()

        await result.current.createRegistry(Config.registry, "user:test")
        await act(async()=>{
            result.current.getRegistries()
        })

        await waitForNextUpdate()
        let found = false
        for(var i=0; i < result.current.data.length; i++) {
            if(result.current.data[i].name === Config.registry) {
                found = true
            }
        }
        expect(found).toBeTrue()

        await result.current.deleteRegistry(Config.registry)

        await act(async()=>{
           result.current.getRegistries()
        })

        await waitForNextUpdate()
        
        found = false
        for(var i=0; i < result.current.data.length; i++) {
            if(result.current.data[i].name === Config.registry) {
                found = true
            }
        }
        expect(found).toBeFalse()
    })
    it('create dumb registry', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useRegistries(Config.url, Config.namespace));
        await waitForNextUpdate()
        let err =    await result.current.createRegistry("not a url", "us e r:tes t")
        console.log(err)
        expect(err).toBe("create registry: Secret \"direktiv-secret-test-\" is invalid: metadata.name: Invalid value: \"direktiv-secret-test-\": a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')")
        // expect(result.current.createErr).not.toBeNull()
    })
    it('delete registry that doesnt exist', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useRegistries(Config.url, Config.namespace));
        await waitForNextUpdate()
        let err=    await result.current.deleteRegistry('test')
        console.log(err)
        expect(err).toBe("delete registry: registry 'test' does not exist")
        // expect(result.current.deleteErr).not.toBeNull()
    })
})


describe('useGlobalRegistries', () => {
    it('list registries',  async () => {
        const { result, waitForNextUpdate } = renderHook(() => useGlobalRegistries(Config.url));
        await waitForNextUpdate()
        
        expect(result.current.data).toBeArray()
    })
    it('create and delete registry', async () => {
        const { result, waitForNextUpdate } = renderHook(() => useGlobalRegistries(Config.url));
        await waitForNextUpdate()

        expect(result.current.data).toBeArray()

        await result.current.createRegistry(Config.registry, "user:test")
        await act(async()=>{
            result.current.getRegistries()
        })

        await waitForNextUpdate()
        let found = false
        for(var i=0; i < result.current.data.length; i++) {
            if(result.current.data[i].name === Config.registry) {
                found = true
            }
        }
        expect(found).toBeTrue()

        await result.current.deleteRegistry(Config.registry)

        await act(async()=>{
           result.current.getRegistries()
        })

        await waitForNextUpdate()
        
        found = false
        for(var i=0; i < result.current.data.length; i++) {
            if(result.current.data[i].name === Config.registry) {
                found = true
            }
        }
        expect(found).toBeFalse()
    })
    it('create dumb registry', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useGlobalRegistries(Config.url));
        await waitForNextUpdate()
        await act(async()=>{
            await result.current.createRegistry("not a url", "us e r:tes t")
        })
        expect(result.current.err).not.toBeNull()
    })
    it('delete registry that doesnt exist', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useGlobalRegistries(Config.url));
        await waitForNextUpdate()
        await act(async()=>{
            await result.current.deleteRegistry('test')
        })
        expect(result.current.err).not.toBeNull()
    })
})

describe('useGlobalPrivateRegistries', () => {
    it('list registries',  async () => {
        const { result, waitForNextUpdate } = renderHook(() => useGlobalPrivateRegistries(Config.url));
        await waitForNextUpdate()
        
        expect(result.current.data).toBeArray()
    })
    it('create and delete registry', async () => {
        const { result, waitForNextUpdate } = renderHook(() => useGlobalPrivateRegistries(Config.url));
        await waitForNextUpdate()

        expect(result.current.data).toBeArray()

        await result.current.createRegistry(Config.registry, "user:test")
        await act(async()=>{
            result.current.getRegistries()
        })

        await waitForNextUpdate()
        let found = false
        for(var i=0; i < result.current.data.length; i++) {
            if(result.current.data[i].name === Config.registry) {
                found = true
            }
        }
        expect(found).toBeTrue()

        await result.current.deleteRegistry(Config.registry)

        await act(async()=>{
           result.current.getRegistries()
        })

        await waitForNextUpdate()
        
        found = false
        for(var i=0; i < result.current.data.length; i++) {
            if(result.current.data[i].name === Config.registry) {
                found = true
            }
        }
        expect(found).toBeFalse()
    })
    it('create dumb registry', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useGlobalPrivateRegistries(Config.url));
        await waitForNextUpdate()
        await act(async()=>{
            await result.current.createRegistry("not a url", "us e r:tes t")
        })
        expect(result.current.err).not.toBeNull()
    })
    it('delete registry that doesnt exist', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useGlobalPrivateRegistries(Config.url));
        await waitForNextUpdate()
        await act(async()=>{
            await result.current.deleteRegistry('test')
        })
        expect(result.current.err).not.toBeNull()
    })
})