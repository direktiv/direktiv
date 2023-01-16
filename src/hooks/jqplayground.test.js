import { renderHook, act } from "@testing-library/react-hooks";
import * as matchers from 'jest-extended';
import { useJQPlayground } from "./index";
import { Config } from "./util";
expect.extend(matchers);

// mock timer using jest
jest.useFakeTimers();

describe('useJQPlayground', () => {
    it('execute a jq command', async ()=> {
        const { result, waitForNextUpdate } = renderHook(() => useJQPlayground(Config.url));
        act(()=>{
            result.current.executeJQ(".test", btoa(JSON.stringify({"test": 2})))
        })
        await waitForNextUpdate()
        expect(result.current.data[0]).toEqual(`${2}`)
    })
    it('execute a bad jq command', async()=>{
        const { result, waitForNextUpdate } = renderHook(() => useJQPlayground(Config.url));
        act(()=>{
            result.current.executeJQ("test2", btoa(JSON.stringify({"test": 2})))
        })
        await waitForNextUpdate()
        expect(result.current.err).not.toBeNull()
    })
})