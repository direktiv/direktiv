import {
  GatewayArray,
  GatewayCheckbox,
  GatewayFilepicker,
  GatewayInput,
  GatewaySelect,
  GatewayTextarea,
} from ".";
import type { Meta, StoryObj } from "@storybook/react";
import { DropdownMenuSeparator } from "../Dropdown";

import { useState } from "react";

const meta = {
  title: "Components/GatewayForms",
  component: GatewayCheckbox,
} satisfies Meta<typeof GatewayCheckbox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => <GatewayCheckbox checked={true}>Check</GatewayCheckbox>,
  argTypes: {},
};

export const BasicAuthFormPlugin = () => {
  const [gwCheckbox1, setgwCheckbox1] = useState(false);
  const [gwCheckbox2, setgwCheckbox2] = useState(false);
  const [gwCheckbox3, setgwCheckbox3] = useState(false);

  return (
    <div className="flex flex-col p-2">
      <GatewayCheckbox
        checked={gwCheckbox1}
        onChange={() => {
          setgwCheckbox1(gwCheckbox1 ? false : true);
        }}
      >
        Add username header:
      </GatewayCheckbox>
      <GatewayCheckbox
        checked={gwCheckbox2}
        onChange={() => {
          setgwCheckbox2(gwCheckbox2 ? false : true);
        }}
      >
        Add tags header:
      </GatewayCheckbox>
      <GatewayCheckbox
        checked={gwCheckbox3}
        onChange={() => {
          setgwCheckbox3(gwCheckbox3 ? false : true);
        }}
      >
        Add groups header:
      </GatewayCheckbox>
    </div>
  );
};

export const KeyAuthFormPlugin = () => {
  const [gwCheckbox1, setgwCheckbox1] = useState(true);
  const [gwCheckbox2, setgwCheckbox2] = useState(true);
  const [gwCheckbox3, setgwCheckbox3] = useState(true);

  return (
    <div className="flex flex-col p-2">
      <GatewayCheckbox
        checked={gwCheckbox1}
        onChange={() => {
          setgwCheckbox1(gwCheckbox1 ? false : true);
        }}
      >
        Add username header:
      </GatewayCheckbox>
      <GatewayCheckbox
        checked={gwCheckbox2}
        onChange={() => {
          setgwCheckbox2(gwCheckbox2 ? false : true);
        }}
      >
        Add tags header:
      </GatewayCheckbox>
      <GatewayCheckbox
        checked={gwCheckbox3}
        onChange={() => {
          setgwCheckbox3(gwCheckbox3 ? false : true);
        }}
      >
        Add groups header:
      </GatewayCheckbox>

      <GatewayInput placeholder="Insert key name">Key name:</GatewayInput>
    </div>
  );
};

export const ACLPlugin = () => {
  const [array, setArray] = useState(["a", "b", "c"]);
  const [array2, setArray2] = useState(() => ["test"]);

  return (
    <div>
      <GatewayArray
        placeholder="insert group name"
        externalArray={array}
        onChange={(changedValue) => {
          setArray(changedValue);
        }}
      >
        Allow Groups:
      </GatewayArray>
      <GatewayArray
        placeholder="insert group name"
        externalArray={array2}
        onChange={(changedValue) => {
          setArray2(changedValue);
        }}
      >
        Deny Groups:
      </GatewayArray>
    </div>
  );
};

export const JSInboundPlugin = () => {
  const [value, setValue] = useState(() => "");

  return (
    <div className="flex flex-col p-2">
      <GatewayTextarea
        value={value}
        onChange={setValue}
        placeholder="Insert Script"
      >
        Script:
      </GatewayTextarea>
    </div>
  );
};

export const JSOutboundPlugin = () => {
  const [value, setValue] = useState(() => "");

  return (
    <div className="flex flex-col p-2">
      <GatewayTextarea
        value={value}
        onChange={setValue}
        placeholder="Insert Script"
      >
        Script:
      </GatewayTextarea>
    </div>
  );
};

export const RequestConverterPlugin = () => {
  const [gwCheckbox1, setgwCheckbox1] = useState(false);
  const [gwCheckbox2, setgwCheckbox2] = useState(false);
  const [gwCheckbox3, setgwCheckbox3] = useState(false);
  const [gwCheckbox4, setgwCheckbox4] = useState(false);

  return (
    <div className="flex flex-col p-2">
      <GatewayCheckbox
        checked={gwCheckbox1}
        onChange={() => {
          setgwCheckbox1(gwCheckbox1 ? false : true);
        }}
      >
        Omit Headers:
      </GatewayCheckbox>

      <GatewayCheckbox
        checked={gwCheckbox2}
        onChange={() => {
          setgwCheckbox2(gwCheckbox2 ? false : true);
        }}
      >
        Omit Queries:
      </GatewayCheckbox>

      <GatewayCheckbox
        checked={gwCheckbox3}
        onChange={() => {
          setgwCheckbox3(gwCheckbox3 ? false : true);
        }}
      >
        Omit Body:
      </GatewayCheckbox>
      <GatewayCheckbox
        checked={gwCheckbox4}
        onChange={() => {
          setgwCheckbox4(gwCheckbox4 ? false : true);
        }}
      >
        Omit Consumer:
      </GatewayCheckbox>
    </div>
  );
};
export const InstantResponse = () => {
  const [value1, setValue1] = useState(() => "200");
  const [value2, setValue2] = useState(() => "");
  const [value3, setValue3] = useState(() => "");

  return (
    <div className="flex flex-col p-2">
      <GatewayInput placeholder="200" value={value1} onChange={setValue1}>
        Status Code:
      </GatewayInput>
      <GatewayInput placeholder="/json" value={value2} onChange={setValue2}>
        Content Type:
      </GatewayInput>
      <GatewayTextarea
        placeholder="Insert Text"
        value={value3}
        onChange={setValue3}
      >
        Status Message:
      </GatewayTextarea>
    </div>
  );
};

export const NamespaceFileTarget = () => {
  const [value, setValue] = useState("");
  const [value2, setValue2] = useState("");
  const [value3, setValue3] = useState("");

  const array = ["Example", "My-Namespace", "Namespace-with-a-very-long-name"];

  return (
    <div className="flex flex-col p-2">
      <GatewaySelect
        data={array}
        placeholder="Select a namespace"
        value={value}
        onValueChange={setValue}
      >
        Namespace:
      </GatewaySelect>
      <GatewayFilepicker
        onChange={setValue2}
        inputValue={value2}
        displayValue={value2}
        placeholder="Type file name"
        buttonText="Choose File"
      >
        File:
      </GatewayFilepicker>
      <GatewayInput value={value3} onChange={setValue3} placeholder="image/jpg">
        Content Type:
      </GatewayInput>
    </div>
  );
};

export const NamespaceVariableTarget = () => {
  const [value, setValue] = useState(() => "");

  const array = ["Example", "My-Namespace", "Namespace-with-a-very-long-name"];

  return (
    <div className="flex flex-col p-2">
      <GatewaySelect
        data={array}
        placeholder="Select a namespace"
        value={value}
        onValueChange={setValue}
      >
        Namespace:
      </GatewaySelect>
      <GatewayInput placeholder="Insert name">Variable:</GatewayInput>
      <GatewayInput placeholder="image/jpg">Content Type:</GatewayInput>
    </div>
  );
};

export const WorkflowVariableTarget = () => {
  const [value1, setValue1] = useState(() => "");
  const [value2, setValue2] = useState(() => "");

  const array1 = ["Example", "My-Namespace", "Namespace-with-a-very-long-name"];
  const array2 = ["a", "b", "c"];

  return (
    <div className="flex flex-col p-2">
      <GatewaySelect
        data={array1}
        placeholder="Select a namespace"
        value={value1}
        onValueChange={setValue1}
      >
        Namespace:
      </GatewaySelect>
      <GatewaySelect
        data={array2}
        placeholder="Select a workflow"
        value={value2}
        onValueChange={setValue2}
      >
        Workflow:
      </GatewaySelect>

      <GatewayInput placeholder="Insert name">Variable:</GatewayInput>
      <GatewayInput placeholder="image/jpg">Content Type:</GatewayInput>
    </div>
  );
};

export const WorkflowTarget = () => {
  const [value1, setValue1] = useState(() => "");
  const [value2, setValue2] = useState(() => "");

  const [gwCheckbox1, setgwCheckbox1] = useState(false);
  const [value3, setValue3] = useState(() => "");

  const array1 = ["Example", "My-Namespace", "Namespace-with-a-very-long-name"];
  const array2 = ["a", "b", "c"];

  return (
    <div className="flex flex-col p-2">
      <GatewaySelect
        data={array1}
        placeholder="Select a namespace"
        value={value1}
        onValueChange={setValue1}
      >
        Namespace:
      </GatewaySelect>
      <GatewaySelect
        data={array2}
        placeholder="Select a workflow"
        value={value2}
        onValueChange={setValue2}
      >
        Workflow:
      </GatewaySelect>
      <GatewayCheckbox
        onChange={() => {
          setgwCheckbox1(gwCheckbox1 ? false : true);
        }}
      >
        Asynchronous:
      </GatewayCheckbox>
      <GatewayInput placeholder="image/jpg" value={value3} onChange={setValue3}>
        Content Type:
      </GatewayInput>
    </div>
  );
};

export const AllFormsFunctionalityDemo = () => {
  const [gwCheckbox, setgwCheckbox] = useState(false);
  const [value, setValue] = useState("");
  const [value1, setValue1] = useState("");
  const array1 = ["Example", "My-Namespace", "Namespace-with-a-very-long-name"];

  const [array, setArray] = useState(["a", "b", "c"]);
  const [value2, setValue2] = useState(() => "");

  const [inputVal, setInputVal] = useState("whatever.png");

  return (
    <div>
      <h3 className="font-bold">Data:</h3>
      <p>Select: {value1}</p>
      <p>Checkbox: {gwCheckbox ? "TRUE" : "FALSE"}</p>
      <p>Input: {value}</p>
      <p>Array: {JSON.stringify(array)}</p>
      <p>Textarea: {value2}</p>
      <p>Filepicker: {inputVal}</p>
      <DropdownMenuSeparator />

      <GatewaySelect
        data={array1}
        placeholder="Select a namespace"
        value={value1}
        onValueChange={setValue1}
      >
        Select:
      </GatewaySelect>
      <GatewayCheckbox
        onChange={() => {
          setgwCheckbox(gwCheckbox ? false : true);
        }}
        checked={gwCheckbox}
      >
        Checkbox:
      </GatewayCheckbox>
      <GatewayInput onChange={setValue} value={value} placeholder="Insert text">
        Input:
      </GatewayInput>
      <GatewayArray
        placeholder="Insert group name"
        externalArray={array}
        onChange={(changedValue) => {
          setArray(changedValue);
        }}
      >
        Array:
      </GatewayArray>
      <GatewayTextarea
        value={value2}
        onChange={setValue2}
        placeholder="Insert script"
      >
        Textarea:
      </GatewayTextarea>
      <GatewayFilepicker
        inputValue={inputVal}
        displayValue={inputVal}
        onChange={setInputVal}
        placeholder="Type file name"
        buttonText="Choose File"
      >
        File:
      </GatewayFilepicker>
    </div>
  );
};
