import { renderHook, act } from "@testing-library/react-hooks";
import * as matchers from 'jest-extended';
import { useEvents } from ".";
import { Config } from "./util";
expect.extend(matchers);

// mock timer using jest
jest.useFakeTimers();


describe('useEvents', ()=>{
    it('stream event listeners', async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useEvents(Config.url, true, Config.namespace))
        await waitForNextUpdate()
        expect(result.current.data).toBeInstanceOf(Array)
    })
    it('fetch event listeners', async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useEvents(Config.url, true, Config.namespace))
        await waitForNextUpdate()
        expect(result.current.data).toBeInstanceOf(Array)
    })
    it('send namespace event', async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useEvents(Config.url, true, Config.namespace))
        await waitForNextUpdate()
        let json = {
            "specversion" : "1.0",
            "type" : "com.github.pull.create",
            "source" : "https://github.com/cloudevents/spec/pull",
            "subject" : "123",
            "id" : "A234-1234-1234",
            "time" : "2018-04-05T17:31:00Z",
            "comexampleextension1" : "value",
            "comexampleothervalue" : 5,
            "datacontenttype" : "text/xml",
            "data" : "<much wow=\"xml\"/>"
        }
        await act(async()=>{
            await result.current.sendEvent(JSON.stringify(json))
        })
        expect(result.current.err).not.toBe(null)
    })
    it('send broken namespace event', async()=>{
        const { result, waitForNextUpdate} = renderHook(()=> useEvents(Config.url, true, Config.namespace))
        await waitForNextUpdate()
        let json = {
            "specversion" : "1.0",
            "type" : "com.github.pull.create",
            "subject" : "123",
            "id" : "A234-1234-1234",
            "time" : "2018-04-05T17:31:00Z",
            "comexampleextension1" : "value",
            "comexampleothervalue" : 5,
            "datacontenttype" : "text/xml",
            "data" : "<much wow=\"xml\"/>"
        }
        await act(async()=>{
            await result.current.sendEvent(JSON.stringify(json))
        })
        expect(result.current.err).toBe("send namespace event: source: REQUIRED\n")
    })
})