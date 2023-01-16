import { renderHook, act } from "@testing-library/react-hooks";
import * as matchers from 'jest-extended';
import {useNamespaceMetrics} from './index'
import { Config } from "./util";
expect.extend(matchers);

// mock timer using jest
jest.useFakeTimers();

describe('useNamespaceMetrics', ()=>{
    it('invoked', async()=>{
        const { result } = renderHook(()=> useNamespaceMetrics(Config.url, Config.namespace))
        let invoked = await result.current.getInvoked()
        expect(invoked.results).toBeInstanceOf(Array)
    })
    it('successful', async()=>{
        const { result } = renderHook(()=> useNamespaceMetrics(Config.url, Config.namespace))
        let successful = await result.current.getSuccessful()
        expect(successful.results).toBeInstanceOf(Array)
    })
    it('failed', async()=>{
        const { result } = renderHook(()=> useNamespaceMetrics(Config.url, Config.namespace))
        let failed = await result.current.getFailed()
        expect(failed.results).toBeInstanceOf(Array)
    })
    it('milliseconds', async()=>{
        const { result } = renderHook(()=> useNamespaceMetrics(Config.url, Config.namespace))
        let milliseconds = await result.current.getMilliseconds()
        expect(milliseconds.results).toBeInstanceOf(Array)
    })
})