import { createFile } from "e2e/utils/files";

export const simpleListenerYaml = `direktiv_api: workflow/v1
description: This workflow spawns an event listener as soon as the file is created
start:
  type: event
  event:
    type: fake.event.one
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!`;

export const contextFiltersListenerYaml = `direktiv_api: workflow/v1
description: This workflow spawns an event listener as soon as the file is created
start:
  type: eventsAnd
  events:
    - type: fake.event.one
      context:
        somekey: somevalue
        more: stuff
    - type: fake.event.two
      context:
        anotherkey: anothervalue
states:
- id: helloworld
  type: noop
  transform:
    result: Hello world!`;

export const createListener = async ({
  name,
  namespace,
  content,
}: {
  name: string;
  namespace: string;
  content: string;
}) => {
  await createFile({
    name,
    namespace,
    type: "workflow",
    content,
  });
};
