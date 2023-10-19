import * as matchers from "jest-extended";

import { act, renderHook } from "@testing-library/react-hooks";

import { Config } from "./util";
import { useBroadcastConfiguration } from "./index";
expect.extend(matchers);

// mock timer using jest
jest.useFakeTimers();

describe("useBroadcastConfiguration", () => {
  it("fetch config", async () => {
    const { result, waitForNextUpdate } = renderHook(() =>
      useBroadcastConfiguration(Config.url, Config.namespace)
    );
    await waitForNextUpdate();
    expect(result.current.data.broadcast["directory.create"]).not.toBeTrue();
  });
  it("set config", async () => {
    const { result, waitForNextUpdate } = renderHook(() =>
      useBroadcastConfiguration(Config.url, Config.namespace)
    );
    await waitForNextUpdate();

    const newdata = result.current.data;
    newdata.broadcast["directory.delete"] = true;

    await act(async () => {
      await result.current.setBroadcastConfiguration(JSON.stringify(newdata));
      await result.current.getBroadcastConfiguration();
    });
    expect(result.current.data.broadcast["directory.delete"]).toBeTrue();
  });
});
