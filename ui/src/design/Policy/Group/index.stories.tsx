import { AndGroup, Connector, OrGroup } from ".";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { Condition } from "../Condition";
import { Placeholder } from "../Placeholder";

const meta = {
  title: "Components/Policy/Group",
  component: AndGroup,
} satisfies Meta<typeof AndGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div className="flex p-10">
      <OrGroup childSizes={[1, 1, 1, 1]}>
        <AndGroup>
          <Condition label="user.email" operator="equal" value="@foobar.org" />
          <Connector />
          <Condition label="user.email" operator="equal" value="@foobar.org" />
          <Connector />
          <Condition label="user.email" operator="equal" value="@foobar.org" />
          <Connector />
          <Placeholder />
        </AndGroup>
        <AndGroup>
          <Condition label="user.email" operator="equal" value="@foobar.org" />
          <Connector />
          <Placeholder />
        </AndGroup>
        <AndGroup>
          <Condition label="user.email" operator="equal" value="@foobar.org" />
          <Connector />
          <Condition label="user.email" operator="equal" value="@foobar.org" />
          <Connector />
          <Placeholder />
        </AndGroup>
        <AndGroup>
          <Placeholder />
        </AndGroup>
      </OrGroup>
    </div>
  ),
  argTypes: {},
};

export const WithNestedORGroup: Story = {
  render: () => (
    <div className="flex p-10">
      <OrGroup childSizes={[1, 1, 2]}>
        <AndGroup>
          <Condition label="user.email" operator="equal" value="@foobar.org" />
          <Connector />
          <Condition label="user.email" operator="equal" value="@foobar.org" />
          <Connector />
          <Condition label="user.email" operator="equal" value="@foobar.org" />
          <Connector />
          <Placeholder />
        </AndGroup>
        <AndGroup>
          <Condition label="user.email" operator="equal" value="@foobar.org" />
          <Connector />
          <Placeholder />
        </AndGroup>
        <AndGroup>
          <Condition label="user.email" operator="equal" value="@foobar.org" />
          <Connector />
          <Placeholder />
          <OrGroup childSizes={[1, 1]}>
            <AndGroup>
              <Condition
                label="user.email"
                operator="equal"
                value="@foobar.org"
              />
              <Connector />
              <Placeholder />
            </AndGroup>
            <AndGroup>
              <Placeholder />
            </AndGroup>
          </OrGroup>
        </AndGroup>
      </OrGroup>
    </div>
  ),
  argTypes: {},
};

export const AlmostBlankSlate: Story = {
  render: () => (
    <div className="flex p-10">
      <AndGroup>
        <Condition label="user.email" operator="equal" value="@foobar.org" />
        <Connector />
        <Placeholder />
      </AndGroup>
    </div>
  ),
  argTypes: {},
};

export const AddAnOrGroup: Story = {
  render: () => (
    <div className="flex p-10">
      <AndGroup>
        <Condition label="user.email" operator="equal" value="@foobar.org" />
        <OrGroup childSizes={[1, 1]}>
          <AndGroup>
            <Placeholder />
          </AndGroup>
          <AndGroup>
            <Placeholder />
          </AndGroup>
        </OrGroup>
        <Placeholder />
      </AndGroup>
    </div>
  ),
  argTypes: {},
};
