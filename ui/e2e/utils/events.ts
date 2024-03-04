import { faker } from "@faker-js/faker";
import { headers } from "./testutils";
import { sendEvent } from "~/api/events/mutate/sendEvent";

/**
 * The mock data is composed so we can predict how many elements should be rendered
 * when testing the filters on the events page. So we have
 * 5 instances for type foo.bar.alpha
 * 4 instances for type foo.bar.beta
 * 6 instances for type foo.bar.gamma
 * 7 instances for type foo.bar.delta
 * 22 instances overall
 * ... and more constellations if we also filter for source.
 *
 * for now, the field "data" (which can also be filtered) is random (and thus unique)
 *
 * Items will not necessarily be listed in this order in the test, because they are
 * created asynchronously.
 */
const eventMockData = [
  { type: "foo.bar.alpha", source: "http://example.one" },
  { type: "foo.bar.alpha", source: "http://example.one" },
  { type: "foo.bar.alpha", source: "http://example.one" },
  { type: "foo.bar.alpha", source: "http://example.two" },
  { type: "foo.bar.alpha", source: "http://example.two" },
  { type: "foo.bar.beta", source: "http://example.one" },
  { type: "foo.bar.beta", source: "http://example.two" },
  { type: "foo.bar.beta", source: "http://example.three" },
  { type: "foo.bar.beta", source: "http://example.three" },
  { type: "foo.bar.gamma", source: "http://example.one" },
  { type: "foo.bar.gamma", source: "http://example.two" },
  { type: "foo.bar.gamma", source: "http://example.two" },
  { type: "foo.bar.gamma", source: "http://example.three" },
  { type: "foo.bar.gamma", source: "http://example.three" },
  { type: "foo.bar.gamma", source: "http://example.three" },
  { type: "foo.bar.delta", source: "http://example.one" },
  { type: "foo.bar.delta", source: "http://example.two" },
  { type: "foo.bar.delta", source: "http://example.two" },
  { type: "foo.bar.delta", source: "http://example.two" },
  { type: "foo.bar.delta", source: "http://example.two" },
  { type: "foo.bar.delta", source: "http://example.three" },
  { type: "foo.bar.delta", source: "http://example.three" },
] as const;

export const createEvents = async (namespace: string) => {
  const events = Array.from(eventMockData, ({ type, source }) => ({
    specversion: "1.0",
    type,
    source,
    data: faker.lorem.sentence(),
    datacontenttype: "text/plain",
  }));

  return await Promise.all(
    events.map((event) =>
      sendEvent({
        payload: event,
        urlParams: {
          baseUrl: process.env.VITE_E2E_UI_DOMAIN,
          namespace,
        },
        headers: {
          ...headers,
          "content-type": "application/cloudevents+json",
        },
        // request returns null, thus return the generated data instead for use in the test
      }).then(() => event)
    )
  );
};
