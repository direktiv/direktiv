import * as matchers from "jest-extended";

import { act, renderHook } from "@testing-library/react-hooks";

import { Config } from "./util";
import { useSecrets } from "./index";
expect.extend(matchers);

// mock timer using jest
// jest.useFakeTimers();

describe("useSecrets", () => {
  it("list secrets", async () => {
    const { result, waitForNextUpdate } = renderHook(() =>
      useSecrets(Config.url, Config.namespace)
    );
    await waitForNextUpdate();

    expect(result.current.data).toBeArray();
  });
  it("create and delete secret", async () => {
    const { result, waitForNextUpdate } = renderHook(() =>
      useSecrets(Config.url, Config.namespace)
    );
    await waitForNextUpdate();
    expect(result.current.data).toBeArray();
    await result.current.createSecret(Config.secret, Config.secretdata);

    await act(async () => {
      result.current.getSecrets();
    });

    await waitForNextUpdate();

    let found = result.current.data.some((x) => x.node.name === Config.secret);
    expect(found).toBeTrue();

    await result.current.deleteSecret(Config.secret);

    await act(async () => {
      result.current.getSecrets();
    });

    await waitForNextUpdate();

    found = result.current.data.some((x) => x.name === Config.secret);
    expect(found).toBeFalse();
  });
  it("create dumb secret", async () => {
    const { result, waitForNextUpdate } = renderHook(() =>
      useSecrets(Config.url, Config.namespace)
    );
    await waitForNextUpdate();
    const err = await result.current.createSecret("not a url", "us e r:tes t");
    expect(err).toBe(
      "create secret: secret name must match the regex pattern `^(([a-z][a-z0-9_\\-]*[a-z0-9])|([a-z]))$`"
    );
  });
  it("delete secret that doesnt exist", async () => {
    const { result, waitForNextUpdate } = renderHook(() =>
      useSecrets(Config.url, Config.namespace)
    );
    await waitForNextUpdate();
    const err = await result.current.deleteSecret("test");
    // todo fix non existent secret
    expect(err).toBe(undefined);
  });
});
