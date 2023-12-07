import {
  GWArray,
  GWCheckbox,
  GWFilepicker,
  GWInput,
  GWSelect,
  GWTextarea,
} from ".";
import type { Meta, StoryObj } from "@storybook/react";
import { DropdownMenuSeparator } from "../Dropdown";

import { useState } from "react";

const meta = {
  title: "Components/GatewayForms",
  component: GWCheckbox,
} satisfies Meta<typeof GWCheckbox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => <GWCheckbox checked={true}>Check</GWCheckbox>,
  argTypes: {},
};

export const BasicAuthFormPlugin = () => {
  const [gwCheckbox1, setgwCheckbox1] = useState(false);
  const [gwCheckbox2, setgwCheckbox2] = useState(false);
  const [gwCheckbox3, setgwCheckbox3] = useState(false);

  return (
    <div className="flex flex-col p-2">
      <GWCheckbox
        checked={gwCheckbox1}
        onChange={() => {
          setgwCheckbox1(gwCheckbox1 ? false : true);
        }}
      >
        Add username header:
      </GWCheckbox>
      <GWCheckbox
        checked={gwCheckbox2}
        onChange={() => {
          setgwCheckbox2(gwCheckbox2 ? false : true);
        }}
      >
        Add tags header:
      </GWCheckbox>
      <GWCheckbox
        checked={gwCheckbox3}
        onChange={() => {
          setgwCheckbox3(gwCheckbox3 ? false : true);
        }}
      >
        Add groups header:
      </GWCheckbox>
    </div>
  );
};

export const KeyAuthFormPlugin = () => {
  const [gwCheckbox1, setgwCheckbox1] = useState(true);
  const [gwCheckbox2, setgwCheckbox2] = useState(true);
  const [gwCheckbox3, setgwCheckbox3] = useState(true);

  return (
    <div className="flex flex-col p-2">
      <GWCheckbox
        checked={gwCheckbox1}
        onChange={() => {
          setgwCheckbox1(gwCheckbox1 ? false : true);
        }}
      >
        Add username header:
      </GWCheckbox>
      <GWCheckbox
        checked={gwCheckbox2}
        onChange={() => {
          setgwCheckbox2(gwCheckbox2 ? false : true);
        }}
      >
        Add tags header:
      </GWCheckbox>
      <GWCheckbox
        checked={gwCheckbox3}
        onChange={() => {
          setgwCheckbox3(gwCheckbox3 ? false : true);
        }}
      >
        Add groups header:
      </GWCheckbox>

      <GWInput placeholder="Insert key name">Key name:</GWInput>
    </div>
  );
};

export const ACLPlugin = () => {
  const [array, setArray] = useState(["a", "b", "c"]);
  const [array2, setArray2] = useState(() => ["test"]);

  return (
    <div>
      <GWArray
        inputPlaceholder="insert group name"
        externalArray={array}
        onChange={(changedValue) => {
          setArray(changedValue);
        }}
      >
        Allow Groups:
      </GWArray>
      <GWArray
        inputPlaceholder="insert group name"
        externalArray={array2}
        onChange={(changedValue) => {
          setArray2(changedValue);
        }}
      >
        Deny Groups:
      </GWArray>
    </div>
  );
};

export const JSInboundPlugin = () => {
  const [value, setValue] = useState(() => "");

  return (
    <div className="flex flex-col p-2">
      <GWTextarea value={value} onChange={setValue} placeholder="Insert Script">
        Script:
      </GWTextarea>
    </div>
  );
};

export const JSOutboundPlugin = () => {
  const [value, setValue] = useState(() => "");

  return (
    <div className="flex flex-col p-2">
      <GWTextarea value={value} onChange={setValue} placeholder="Insert Script">
        Script:
      </GWTextarea>
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
      <GWCheckbox
        checked={gwCheckbox1}
        onChange={() => {
          setgwCheckbox1(gwCheckbox1 ? false : true);
        }}
      >
        Omit Headers:
      </GWCheckbox>

      <GWCheckbox
        checked={gwCheckbox2}
        onChange={() => {
          setgwCheckbox2(gwCheckbox2 ? false : true);
        }}
      >
        Omit Queries:
      </GWCheckbox>

      <GWCheckbox
        checked={gwCheckbox3}
        onChange={() => {
          setgwCheckbox3(gwCheckbox3 ? false : true);
        }}
      >
        Omit Body:
      </GWCheckbox>
      <GWCheckbox
        checked={gwCheckbox4}
        onChange={() => {
          setgwCheckbox4(gwCheckbox4 ? false : true);
        }}
      >
        Omit Consumer:
      </GWCheckbox>
    </div>
  );
};
export const InstantResponse = () => {
  const [value1, setValue1] = useState(() => "200");
  const [value2, setValue2] = useState(() => "");
  const [value3, setValue3] = useState(() => "");

  return (
    <div className="flex flex-col p-2">
      <GWInput placeholder="200" value={value1} onChange={setValue1}>
        Status Code:
      </GWInput>
      <GWInput placeholder="/json" value={value2} onChange={setValue2}>
        Content Type:
      </GWInput>
      <GWTextarea placeholder="Insert Text" value={value3} onChange={setValue3}>
        Status Message:
      </GWTextarea>
    </div>
  );
};

export const NamespaceFileTarget = () => {
  const [value, setValue] = useState(() => "");

  const array = ["Example", "My-Namespace", "Namespace-with-a-very-long-name"];

  return (
    <div className="flex flex-col p-2">
      <GWSelect
        data={array}
        placeholder="Select a namespace"
        value={value}
        onValueChange={setValue}
      >
        Namespace:
      </GWSelect>
      <GWFilepicker placeholder="Type file name" buttonText="Choose File">
        File:
      </GWFilepicker>
      <GWInput placeholder="image/jpg">Content Type:</GWInput>
    </div>
  );
};

export const NamespaceVariableTarget = () => {
  const [value, setValue] = useState(() => "");

  const array = ["Example", "My-Namespace", "Namespace-with-a-very-long-name"];

  return (
    <div className="flex flex-col p-2">
      <GWSelect
        data={array}
        placeholder="Select a namespace"
        value={value}
        onValueChange={setValue}
      >
        Namespace:
      </GWSelect>
      <GWInput placeholder="Insert name">Variable:</GWInput>
      <GWInput placeholder="image/jpg">Content Type:</GWInput>
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
      <GWSelect
        data={array1}
        placeholder="Select a namespace"
        value={value1}
        onValueChange={setValue1}
      >
        Namespace:
      </GWSelect>
      <GWSelect
        data={array2}
        placeholder="Select a workflow"
        value={value2}
        onValueChange={setValue2}
      >
        Workflow:
      </GWSelect>

      <GWInput placeholder="Insert name">Variable:</GWInput>
      <GWInput placeholder="image/jpg">Content Type:</GWInput>
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
      <GWSelect
        data={array1}
        placeholder="Select a namespace"
        value={value1}
        onValueChange={setValue1}
      >
        Namespace:
      </GWSelect>
      <GWSelect
        data={array2}
        placeholder="Select a workflow"
        value={value2}
        onValueChange={setValue2}
      >
        Workflow:
      </GWSelect>
      <GWCheckbox
        onChange={() => {
          setgwCheckbox1(gwCheckbox1 ? false : true);
        }}
      >
        Asynchronous:
      </GWCheckbox>
      <GWInput placeholder="image/jpg" value={value3} onChange={setValue3}>
        Content Type:
      </GWInput>
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
  const [displayVal, setDisplayVal] = useState(inputVal);

  return (
    <div>
      <h3 className="font-bold">Data:</h3>
      <p>Select: {value1}</p>
      <p>Checkbox: {gwCheckbox ? "TRUE" : "FALSE"}</p>
      <p>Input: {value}</p>
      <p>Array: {JSON.stringify(array)}</p>
      <p>Textarea: {value2}</p>
      <p>Filepicker: {displayVal}</p>
      <DropdownMenuSeparator />

      <GWSelect
        data={array1}
        placeholder="Select a namespace"
        value={value1}
        onValueChange={setValue1}
      >
        Select:
      </GWSelect>
      <GWCheckbox
        onChange={() => {
          setgwCheckbox(gwCheckbox ? false : true);
        }}
        checked={gwCheckbox}
      >
        Checkbox:
      </GWCheckbox>
      <GWInput onChange={setValue} value={value} placeholder="Insert text">
        Input:
      </GWInput>
      <GWArray
        inputPlaceholder="Insert group name"
        externalArray={array}
        onChange={(changedValue) => {
          setArray(changedValue);
        }}
      >
        Array:
      </GWArray>
      <GWTextarea
        value={value2}
        onChange={setValue2}
        placeholder="Insert script"
      >
        Textarea:
      </GWTextarea>
      <GWFilepicker
        onClick={() => {
          setDisplayVal(inputVal);
        }}
        inputValue={inputVal}
        displayValue={displayVal}
        onChange={setInputVal}
        placeholder="Type file name"
        buttonText="Choose File"
      >
        File:
      </GWFilepicker>
    </div>
  );
};
