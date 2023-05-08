import { Card } from "../Card";
import Editor from "./index";

import type { Meta } from "@storybook/react";

export default {
  title: "Components/Editor",
} satisfies Meta<typeof Editor>;

const value = `# some comment here
functions:
- id: greeter
  image: direktiv/greeting:v3
  type: knative-workflow
- id: solve2
  image: direktiv/solve:v3
  type: knative-workflow
description: A simple 'eventXor' that waits for events to be received.
states:
- id: event-xor
  type: eventXor
  timeout: PT1H
  events:
  - event: 
      type: solveexpressioncloudevent
    transition: solve
  - event: 
      type: greetingcloudevent
    transition: greet
- id: greet
  type: action
  action:
    function: greeter
    input: jq(.greetingcloudevent.data)
  transform: 
    greeting: jq(.return.greeting)
- id: solve
  type: action
  action:
    function: solve2
    input: jq(.solveexpressioncloudevent.data)
  transform: 
    solvedexpression: jq(.return)
`;

export const Default = () => (
  <div className="flex flex-col gap-y-3 bg-white p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="h-[500px]">
      <Editor value={value} />
    </div>
  </div>
);

export const Small = () => (
  <div className="flex flex-col gap-y-3 bg-white p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="h-[500px] w-[500px]">
      <Editor value={value} />
    </div>
  </div>
);
export const Darkmode = () => (
  <div className="flex flex-col gap-y-3 bg-black p-4">
    <div>This Story is not aware of light and dark mode.</div>
    <div className="h-[500px]">
      <Editor value={value} theme="dark" />
    </div>
  </div>
);

export const WithCardAnd100Height = () => (
  <div className="flex h-[97vh] min-h-full flex-col gap-y-3 bg-black">
    <div>This Story is not aware of light and dark mode.</div>
    <Card className="grow p-4">
      <Editor value={value} theme="dark" />
    </Card>
  </div>
);
