import EventSource from "eventsource";
import common from "../common";

const createNamespaceResponse = {
  namespace: common.structs.namespaceObject,
};

const API_HOST = "http://ec2-3-231-218-167.compute-1.amazonaws.com";
const AUTH_TOKEN = "password"; // can also be null if no password should be used
const IS_ENTERPRISE = false; // enterprise needs different auth header
const NAMESPACE = "my-namespace";

const authHeader = (authToken, isEnterprice = false) => {
  if (!authToken) {
    return {};
  }
  return isEnterprice
    ? { Authorization: `Bearer ${authToken}` }
    : { "direktiv-token": authToken };
};

describe("Test Server Sent Events", () => {
  // write a handler
  it(`can handle Server Sent Events`, async () => {
    const sseListener = new EventSource(
      `${API_HOST}/api/namespaces/${NAMESPACE}/logs`,
      {
        headers: {
          ...authHeader(AUTH_TOKEN, IS_ENTERPRISE),
        },
      }
    );

    // we need to wait for a promise here, otherwise the test will finish before the event is received
    await new Promise((resolve, reject) => {
      sseListener.onmessage = (e) => {
        const json = JSON.parse(e.data);
        expect(json).toMatchObject({
          pageInfo: {
            total: json.results.length,
          },
          namespace: NAMESPACE,
        });

        sseListener.close();
        resolve();
      };

      // rejecting a promise in jest will fail the test
      sseListener.onerror = (e) => {
        reject(e);
      };
    });
  });
});
