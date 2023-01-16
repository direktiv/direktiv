import { renderHook, act } from "@testing-library/react-hooks";
import * as matchers from 'jest-extended';
import {useBroadcastConfiguration} from './index'
import { Config } from "./util";
expect.extend(matchers);

// mock timer using jest
jest.useFakeTimers();

describe('useBroadcastConfiguration', ()=>{
    it('fetch config', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useBroadcastConfiguration(Config.url, Config.namespace));
        await waitForNextUpdate()
        expect(result.current.data.broadcast["directory.create"]).not.toBeTrue()
    })
    it('set config', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useBroadcastConfiguration(Config.url, Config.namespace));
        await waitForNextUpdate()

        let newdata = result.current.data
        newdata.broadcast['directory.delete'] = true

        await act(async ()=>{
            await result.current.setBroadcastConfiguration(JSON.stringify(newdata))
            await result.current.getBroadcastConfiguration()
        })
        expect(result.current.data.broadcast["directory.delete"]).toBeTrue()
    })
})