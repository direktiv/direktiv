import { LogEntry, Logs } from "./index";
import type { Meta, StoryObj } from "@storybook/react";
import { Card } from "../Card";

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
        <Logs>
            <LogEntry time="12:34:23" linewrap>
                Hey this is the error log Hey this is the error logHey this is the error
                logHey this is the error logHey this is the error logHey this is the
                error logHey this is the error logHey this is the error logHey this is
                the error logHey this is the error logHey this is the error logHey this
                is the error logHey this is the error logHey this is the error logHey
                this is the error logHey this is the error logHey this is the error
                logHey this is the error logHey this is the error logHey this is the
                error logHey this is the error logHey this is the error logHey this is
                the error log
            </LogEntry>
        </Logs>
        <Logs>
            <LogEntry time="12:34:23">
                Hey this is the error log Hey this is the error logHey this is the error
                logHey this is the error logHey this is the error logHey this is the
                error logHey this is the error logHey this is the error logHey this is
                the error logHey this is the error logHey this is the error logHey this
                is the error logHey this is the error logHey this is the error logHey
                this is the error logHey this is the error logHey this is the error
                logHey this is the error logHey this is the error logHey this is the
                error logHey this is the error logHey this is the error logHey this is
                the error log
            </LogEntry>
        </Logs>
    </Card>
);
