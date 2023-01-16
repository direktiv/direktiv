import * as matchers from 'jest-extended';
import { HandleError } from './util';
expect.extend(matchers);

// mock timer using jest
jest.useFakeTimers();

describe('utilHandleError', () => {
    it('test forbidden', async ()=> {
        let resp = {
            headers: new Map().set("content-type","application/json"),
            status: 403
        }
        expect.toBeString(await HandleError('test forbidden', resp, 'forbidden'))
    })
    it('test method not allowed', async ()=> {
        let resp = {
            headers: new Map().set("content-type","application/json"),
            status: 405
        }
        expect.toBeString(await HandleError('test forbidden', resp, 'forbidden'))
    })
    it('test grpc message', async ()=> {
        let resp = {
            headers: new Map().set("content-type","application/json").set("grpc-message", "an error has occurred"),
            status: 400
        }
        expect.toBeString(await HandleError('test forbidden', resp, 'forbidden'))
    })
    it('test non json result', async ()=> {
        let resp = {
            headers: new Map().set("content-type","plain/text"),
            status: 400,
            text: ()=>{return "an error has happened"}
        }
        expect.toBeString(await HandleError('test forbidden', resp, 'forbidden'))
    })
    it('test non json result', async ()=> {
        let resp = {
            headers: new Map().set("content-type","application/json"),
            status: 400,
            json: ()=>{return {message: "error has occurred"}}
        }
        expect.toBeString(await HandleError('test forbidden', resp, 'forbidden'))
    })
})