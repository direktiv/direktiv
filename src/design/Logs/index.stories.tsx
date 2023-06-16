import { LogEntry, Logs } from "./index";
import type { Meta, StoryObj } from "@storybook/react";
import { Card } from "../Card";
import Editor from "../Editor";
import { YamlSample } from "../Editor/languageSamples";

const meta = {
  title: "Components/Logs",
  component: Logs,
} satisfies Meta<typeof Logs>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <Logs {...args}>
      <LogEntry time="12:342:23" variant="error">
        Hey this is the log
      </LogEntry>

      <LogEntry time="12:342:23" variant="error">
        Hey this is the log
      </LogEntry>
    </Logs>
  ),
};

export const LogVariants = () => (
  <Logs>
    <LogEntry time="12:34:23" variant="error">
      Hey this is the error log
    </LogEntry>

    <LogEntry time="12:34:23" variant="success">
      Hey this is the success log
    </LogEntry>

    <LogEntry time="12:34:23" variant="warning">
      Hey this is the warning log
    </LogEntry>
    <LogEntry time="12:34:23" variant="info">
      Hey this is the info log
    </LogEntry>
    <LogEntry time="12:34:23">Hey this is the info log</LogEntry>
  </Logs>
);

export const WrapLog = () => (
  <Card>
    <Logs linewrap>
      <LogEntry time="12:34:23" variant="success">
        {`This is going to be a very long line This is going to be a very long line This is going to be a very long line This is going to be a very long line This is going to be a very long line
          Next line
          Third line`}
      </LogEntry>
    </Logs>
  </Card>
);

export const NoWrapLog = () => (
  <Card>
    <Logs>
      <LogEntry time="12:34:23" variant="success">
        {`This is going to be a very long line This is going to be a very long line This is going to be a very long line This is going to be a very long line This is going to be a very long line
          Next line
          Third line
        `}
      </LogEntry>
      <LogEntry time="12:34:23" variant="error">
        New Entry
      </LogEntry>
    </Logs>
  </Card>
);

export const EditorVSLogsFontCompare = () => (
  <div className="flex flex-row gap-5">
    <Card>
      <Logs>
        <LogEntry time="12:34:23" variant="success">
          {`${YamlSample}
          `}
        </LogEntry>
      </Logs>
    </Card>
    <Card>
      <div className="h-[500px] w-[500px]">
        <Editor value={YamlSample} />
      </div>
    </Card>
  </div>
);
