import { renderHook, act } from "@testing-library/react-hooks";
import * as matchers from 'jest-extended';
import { Config } from './util';
import { useNamespaceLogs, useNamespaces } from './index';
expect.extend(matchers);

// mock timer using jest
jest.useFakeTimers();
jest.setTimeout(30000)

// describe('useNamespaceLogs', ()=> {
//     it('fetch namespace logs', async() => {
//         const { result, waitForNextUpdate } = renderHook(() => useNamespaceLogs(Config.url, false, Config.namespace));
//         await waitForNextUpdate()
//         expect(result.current.data).toBeArray()
//     })
//     it('fetch namespace logs then fetch again', async()=>{
//         const { result, waitForNextUpdate } = renderHook(() => useNamespaceLogs(Config.url, false, Config.namespace));
//         await waitForNextUpdate()
//         expect(result.current.data).toBeArray()
//         act(()=>{
//             result.current.getNamespaceLogs()
//         })
//         await waitForNextUpdate()
//     })
//     it('stream namespace logs', async()=> {
//         const { result, waitForNextUpdate } = renderHook(() => useNamespaceLogs(Config.url, true, Config.namespace));
//         await waitForNextUpdate()
//         expect(result.current.data).toBeArray()
//     })
// })

describe('useNamespaces', () => {
//   it('fetch namespaces',  async () => {
//     const { result, waitForNextUpdate } = renderHook(() => useNamespaces(Config.url, false));
//     await waitForNextUpdate()
//     expect(result.current.data).toBeArray()
//   })
  it('stream namespaces', async() => {
    const { result, waitForNextUpdate } = renderHook(() => useNamespaces(Config.url, true));
    console.log(result.current)
    await waitForNextUpdate()
    console.log(result.current)
    expect(result.current.data).toBeArray()
  })
//   it('fetch namespaces with apikey', async() => {
//     const { result, waitForNextUpdate } = renderHook(() => useNamespaces(Config.url, false, Config.apikey));
//     await waitForNextUpdate()
//     expect(result.current.data).toBeArray()
//   })
  // TODO cant test because of cors
//   it('stream namespaces with apikey', async() => {
//     const { result, waitForNextUpdate } = renderHook(() => useNamespaces(Config.url, true, Config.apikey));
//     await waitForNextUpdate()
//     expect(result.current.data).toBeArray()
//   })
//   it('broken url error', async() => {
//     const { result, waitForNextUpdate } = renderHook(() => useNamespaces(Config.url+"x", false));
//     await waitForNextUpdate()
//     expect (result.current.err).not.toBeNull()
//   })
//   it('create a namespace with apikey', async()=>{
//     const { result, waitForNextUpdate } = renderHook(() => useNamespaces(Config.url, false,Config.apikey));
//     // wait for initial result    
//     await waitForNextUpdate()

//     await result.current.createNamespace("test-test")

//     await act(async()=>{
//         result.current.getNamespaces()
//     })

//     await waitForNextUpdate()
//     let found = false
//     for(var i=0; i < result.current.data.length; i++) {
//         if(result.current.data[i].node.name === "test-test") {
//             found = true
//         }
//     }
//     expect(found).toBeTrue()
//   })
//   it('delete a namespace with apikey', async()=>{
//     const { result, waitForNextUpdate } = renderHook(() => useNamespaces(Config.url, false, Config.apikey));
//     // wait for initial result
//     await waitForNextUpdate()
    
//     // delete a namespace
//     await result.current.deleteNamespace("test-test")

//     await act(async()=>{
//         result.current.getNamespaces()
//     })

//     await waitForNextUpdate()
//     let found = false
//     for(var i=0; i < result.current.data.length; i++) {
//         if(result.current.data[i].node.name === "test-test") {
//             found = true
//         }
//     }
//     expect(found).toBeFalse()
//   })
//   it('create a namespace', async() => {
//     const { result, waitForNextUpdate } = renderHook(() => useNamespaces(Config.url, true));
//     // wait for initial result    
//     await waitForNextUpdate()

//     // create a namespace
// await        result.current.createNamespace("test-test")

// await waitForNextUpdate()

//     let found = false
//     for(var i=0; i < result.current.data.length; i++) {
//         if(result.current.data[i].node.name === "test-test") {
//             found = true
//         }
//     }
//     expect(found).toBeTrue()
//   })
//   it('delete a namespace', async() => {
//     const { result, waitForNextUpdate } = renderHook(() => useNamespaces(Config.url, true));
//     // wait for initial result
//     await waitForNextUpdate()
    
//     // delete a namespace
//         await result.current.deleteNamespace("test-test")
//     await waitForNextUpdate()
//     let found = false
//     for(var i=0; i < result.current.data.length; i++) {
//         if(result.current.data[i].node.name === "test-test") {
//             found = true
//         }
//     }
//     expect(found).toBeFalse()
//   })
//   it('delete a namespace that doesnt exist', async()=>{
//     const { result, waitForNextUpdate } = renderHook(() => useNamespaces(Config.url, true));
//     await waitForNextUpdate()
//     let err = await result.current.deleteNamespace("xxxxxx")
//     expect(err).toBe('delete a namespace: namespace not found')
//   })
//   it('create a namespace with a bad name', async()=>{
//     const { result, waitForNextUpdate } = renderHook(() => useNamespaces(Config.url, true));
//     await waitForNextUpdate()
//     let err =  await result.current.createNamespace("Xxxxxx")
//     expect(err).toBe("create a namespace: one or more fields has an invalid value")
//   })
})
