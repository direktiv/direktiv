import { renderHook, act } from "@testing-library/react-hooks";
import * as matchers from 'jest-extended';
import { Config } from "./util";
import { useWorkflow, useWorkflowServiceRevision, useWorkflowService, useGlobalServiceRevision,  usePodLogs, useGlobalServices, useNamespaceServices, useNodes, useWorkflowServices, useNamespaceServiceRevision, useNamespaceService} from './index'
expect.extend(matchers);

// mock timer using jest
jest.useFakeTimers();
jest.setTimeout(70000)

const wfyaml = `functions:
- id: req
  image: direktiv/request:v1
  type: reusable
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!
`

// execute wf to trigger service to be created
async function executeWF() {
    console.log('execute workflow')
    const { result, waitForNextUpdate} = renderHook(()=> useWorkflow(Config.url, true, Config.namespace, "test-workflow-services"))
    await waitForNextUpdate()
    result.current.executeWorkflow()
}

beforeAll(async ()=>{
    console.log('creating workflow')
    const { result, waitForNextUpdate} = renderHook(()=> useNodes(Config.url, true, Config.namespace, ""))
    await waitForNextUpdate()
    await act( async()=>{
        await result.current.createNode("/test-workflow-services", "workflow",  wfyaml)
        await executeWF()
    })
})

describe('useWorkflowServices',()=>{
    it('fetch workflow services', async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useWorkflowServices(Config.url, false, Config.namespace, "test-workflow-services"))
        await waitForNextUpdate()
        expect(result.current.data.config.maxscale).toBe(3)
    })
    it('stream workflow services', async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useWorkflowServices(Config.url, true, Config.namespace, "test-workflow-services"))
        await waitForNextUpdate()
        expect(result.current.data[0].info.path).toBe("test-workflow-services")
        process.env.WORKFLOW_REVISION = result.current.data[0].info.revision
        process.env.WORKFLOW_SERVICE_NAME = result.current.data[0].info.name
    })
    describe('useWorkflowService', ()=>{
        it('fetch service details (fn, revisions, traffic)', async()=>{
            const { result, waitForNextUpdate } = renderHook(()=> useWorkflowService(Config.url, Config.namespace, "test-workflow-services", process.env.WORKFLOW_SERVICE_NAME, process.env.WORKFLOW_REVISION))
            await waitForNextUpdate()
            expect(result.current.revisions[0].image).toBe("direktiv/request:v1")
            process.env.WORKFLOW_REV = result.current.revisions[0].rev
        })
        describe('useWorkflowServiceRevision', ()=>{
            it('get revision details and pods', async()=>{
                const { result, waitForNextUpdate} = renderHook(()=> useWorkflowServiceRevision(Config.url, Config.namespace, "test-workflow-services", process.env.WORKFLOW_SERVICE_NAME, process.env.WORKFLOW_REVISION, process.env.WORKFLOW_REV))
                await waitForNextUpdate()
                await waitForNextUpdate()
                await waitForNextUpdate()
                expect(result.current.revisionDetails.image).toBe("direktiv/request:v1")
                expect(result.current.pods).toBeInstanceOf(Array)
                process.env.WF_POD_ID = result.current.pods[0].name
            })
            describe('usePodLogs', ()=>{
                it('get pods logs', async()=>{
                    const { result, waitForNextUpdate} = renderHook(()=> usePodLogs(Config.url, process.env.WF_POD_ID))
                    await waitForNextUpdate()
                    expect(result.current.data.data).toBe("Starting server.\n")
                })
            })
        })
    })
})

describe('useNamespaceServices', ()=>{
    it('fetch namespace services', async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useNamespaceServices(Config.url, false, Config.namespace))
        await waitForNextUpdate()
        expect(result.current.data.config.maxscale).toBe(3)
    })
    it('stream namespace services, create a service then delete a service', async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useNamespaceServices(Config.url, true, Config.namespace))

        await act(async()=>{
            await result.current.createNamespaceService('testglobalservice', 'direktiv/request:v1', 0, 0, "")
            await result.current.createNamespaceService('service-revision-test-global', 'direktiv/request:v1', 0, 0, "")
        })

        let found = false
        for(var i=0; i < result.current.data.length; i++) {
            if(result.current.data[i].info.name === "testglobalservice") {
                found = true
            }
        }
        expect(found).toBeTrue()
    
        await act(async()=>{
            await result.current.deleteNamespaceService('testglobalservice')
        })

        found = false
        for(var i=0; i < result.current.data.length; i++) {
            if(result.current.data[i].info.name === "testglobalservice") {
                found = true
            }
        }
        expect(found).not.toBeTrue()
    })
    describe('useNamespaceService', ()=>{
        it('fetch service details (fn, revisions, traffic)', async()=>{
            const { result, waitForNextUpdate} = renderHook(()=> useNamespaceService(Config.url, Config.namespace, "service-revision-test-global"))
            await waitForNextUpdate()
            await waitForNextUpdate()
            expect(result.current.fn.info.name).toBe("service-revision-test-global")
            expect(result.current.revisions[0].image).toBe("direktiv/request:v1")
            expect(result.current.traffic).toBeInstanceOf(Array)
        })
        it('create a new revision, set traffic to 50/50 then delete the first revision', async()=>{
            const { result, waitForNextUpdate} = renderHook(()=> useNamespaceService(Config.url, Config.namespace,"service-revision-test-global"))
            await waitForNextUpdate()

            await act(async()=>{
                await result.current.createNamespaceServiceRevision('direktiv/request:v1', 0, 0, "", 100)
            })

            await waitForNextUpdate()

            expect(result.current.revisions.length).toBe(2)

            // set traffic to 50/50
            await act(async()=>{
                await result.current.setNamespaceServiceRevisionTraffic(result.current.revisions[0].name, 50, result.current.revisions[1].name, 50)
            })

            await waitForNextUpdate()

            await act(async()=>{
                // delete 1st revision
                await result.current.deleteNamespaceServiceRevision("00001")
            })

            await waitForNextUpdate()

            expect(result.current.revisions.length).toBe(1)
        })
        describe('useNamespaceServiceRevision', ()=>{
            it('get revision details and pods', async()=>{
                const { result, waitForNextUpdate} = renderHook(()=> useNamespaceServiceRevision(Config.url, Config.namespace, "service-revision-test-global", "00002"))
                await waitForNextUpdate()
                await waitForNextUpdate()
                await waitForNextUpdate()
                expect(result.current.revisionDetails.image).toBe("direktiv/request:v1")
                expect(result.current.pods).toBeInstanceOf(Array)
                process.env.NS_POD_ID = result.current.pods[0].name
            })
        })
        describe('usePodLogs', ()=>{
            it('get pods logs', async()=>{
                const { result, waitForNextUpdate} = renderHook(()=> usePodLogs(Config.url, process.env.NS_POD_ID))
                await waitForNextUpdate()
                expect(result.current.data.data).toBe("Starting server.\n")
            })
        })
    })    
})

describe('useGlobalServices', ()=>{
    it('fetch global services', async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useGlobalServices(Config.url, false))
        await waitForNextUpdate()
        expect(result.current.data.config.maxscale).toBe(3)
    })
    it('stream global services, create a service then delete a service', async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useGlobalServices(Config.url, true))

        await act(async()=>{
            await result.current.createGlobalService('testglobalservice', 'direktiv/request:v1', 0, 0, "")
            await result.current.createGlobalService('service-revision-test-global', 'direktiv/request:v1', 0, 0, "")
        })

        let found = false
        for(var i=0; i < result.current.data.length; i++) {
            if(result.current.data[i].info.name === "testglobalservice") {
                found = true
            }
        }
        expect(found).toBeTrue()
    
        await act(async()=>{
            await result.current.deleteGlobalService('testglobalservice')
        })

        found = false
        for(var i=0; i < result.current.data.length; i++) {
            if(result.current.data[i].info.name === "testglobalservice") {
                found = true
            }
        }
        expect(found).not.toBeTrue()
    })
    describe('useGlobalService', ()=>{
        it('fetch service details (fn, revisions, traffic)', async()=>{
            const { result, waitForNextUpdate} = renderHook(()=> useGlobalService(Config.url, "service-revision-test-global"))
            await waitForNextUpdate()
            await waitForNextUpdate()
            expect(result.current.fn.info.name).toBe("service-revision-test-global")
            expect(result.current.revisions[0].image).toBe("direktiv/request:v1")
            expect(result.current.traffic).toBeInstanceOf(Array)
        })
        it('create a new revision, set traffic to 50/50 then delete the first revision', async()=>{
            const { result, waitForNextUpdate} = renderHook(()=> useGlobalService(Config.url, "service-revision-test-global"))
            await waitForNextUpdate()

            await act(async()=>{
                await result.current.createGlobalServiceRevision('direktiv/request:v1', 0, 0, "", 100)
            })

            await waitForNextUpdate()

            expect(result.current.revisions.length).toBe(2)

            // set traffic to 50/50
            await act(async()=>{
                await result.current.setGlobalServiceRevisionTraffic(result.current.revisions[0].name, 50, result.current.revisions[1].name, 50)
            })

            await waitForNextUpdate()

            await act(async()=>{
                // delete 1st revision
                await result.current.deleteGlobalServiceRevision("00001")
            })

            await waitForNextUpdate()

            expect(result.current.revisions.length).toBe(1)
        })
        describe('useGlobalServiceRevision', ()=>{
            it('get revision details and pods', async()=>{
                const { result, waitForNextUpdate} = renderHook(()=> useGlobalServiceRevision(Config.url, "service-revision-test-global", "00002"))
                await waitForNextUpdate()
                await waitForNextUpdate()
                await waitForNextUpdate()
                expect(result.current.revisionDetails.image).toBe("direktiv/request:v1")
                expect(result.current.pods).toBeInstanceOf(Array)
                process.env.POD_ID = result.current.pods[0].name
            })
        })
        describe('usePodLogs', ()=>{
            it('get pods logs', async()=>{
                const { result, waitForNextUpdate} = renderHook(()=> usePodLogs(Config.url, process.env.POD_ID))
                await waitForNextUpdate()
                expect(result.current.data.data).toBe("Starting server.\n")
            })
        })
    })    
})

afterAll(async ()=>{
    console.log('deleting workflow')

    const { result, waitForNextUpdate} = renderHook(()=> useNodes(Config.url, true, Config.namespace, ""))
    await waitForNextUpdate()
    act( ()=>{
        result.current.deleteNode('test-workflow-services')
    })
    await waitForNextUpdate()
})