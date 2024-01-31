import type { Meta, StoryObj } from "@storybook/react";
import {
  Variablepicker,
  VariablepickerError,
  VariablepickerHeading,
  VariablepickerItem,
  VariablepickerMessage,
  VariablepickerSeparator,
} from "./";

import { ButtonBar } from "../ButtonBar";
import Input from "../Input";
import { useState } from "react";

const meta = {
  title: "Components/Variablepicker",
  component: Variablepicker,
} satisfies Meta<typeof Variablepicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <Variablepicker {...args}>
      <VariablepickerItem value="File1">File1</VariablepickerItem>
      <VariablepickerItem value="File2">File2</VariablepickerItem>
      <VariablepickerItem value="File3">File3</VariablepickerItem>
    </Variablepicker>
  ),
  args: {
    buttonText: "Browse Files",
  },
  argTypes: {},
};

const variableList: string[] = [
  "image.jpg",
  "hello.yaml",
  "hello1.yaml",
  "Readme.txt",
  "Readme0.txt",
];

export const WithInputField = () => {
  const defaultValue = "defaultValue";
  const [inputValue, setInputValue] = useState(
    defaultValue ? defaultValue : ""
  );
  const buttonText = "Browse Variables";
  return (
    <ButtonBar>
      <Variablepicker
        buttonText={buttonText}
        onValueChange={(variable) => {
          setInputValue(variable);
        }}
      >
        {variableList.map((variable, index) => (
          <VariablepickerItem key={index} value={variable}>
            {variable}
          </VariablepickerItem>
        ))}
      </Variablepicker>
      <Input
        placeholder="Select a Variable"
        value={inputValue}
        onChange={(e) => {
          setInputValue(e.target.value);
        }}
      />
    </ButtonBar>
  );
};

export const WithHeadingAndSeparator = () => {
  const [value, setValue] = useState("");

  return (
    <div className="flex items-center ">
      <Variablepicker
        value={value}
        onValueChange={(value) => {
          setValue(value);
        }}
        buttonText="Select Variable"
      >
        <VariablepickerHeading>Your Variables:</VariablepickerHeading>
        <VariablepickerSeparator />
        <VariablepickerItem value="one">one</VariablepickerItem>
        <VariablepickerSeparator />
        <VariablepickerItem value="two">two</VariablepickerItem>
        <VariablepickerSeparator />
        <VariablepickerItem value="three">three</VariablepickerItem>
      </Variablepicker>
    </div>
  );
};

export const WithErrorMessage = () => (
  <>
    <p>
      <b>Explanation:</b>
      <br /> For errors in the Variablepicker we use a popover element because a
      select element with one item does not make sense, and also we do not need
      to select it or deliver the value somewhere.
    </p>
    <br />
    <div className="flex items-center ">
      <VariablepickerError buttonText="Select Variable">
        <VariablepickerHeading>Variables:</VariablepickerHeading>
        <VariablepickerSeparator />
        <VariablepickerMessage>- This space is empty -</VariablepickerMessage>
        <VariablepickerSeparator />
      </VariablepickerError>
    </div>
  </>
);
