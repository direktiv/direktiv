import type { Meta, StoryObj } from "@storybook/react";
import {
  Variablepicker,
  VariablepickerHeading,
  VariablepickerItem,
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

export const WithMappingItems = () => {
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
  const [inputValue, setInputValue] = useState("");

  return (
    <div className="flex items-center ">
      <Variablepicker
        onValueChange={(value) => {
          setInputValue(value);
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
      <div className="m-5 flex">
        <p>It is:</p>
        <p className="ml-2">{inputValue}</p>
      </div>
    </div>
  );
};
