import {
  ArrowRight,
  File,
  FolderUp,
  LucideIcon,
  MenuSquareIcon,
  Plus,
  SettingsIcon,
  X,
} from "lucide-react";

import { Breadcrumb, BreadcrumbRoot } from "../Breadcrumbs";
import { Command, CommandGroup, CommandList } from "../Command";

import {
  Filepicker,
  FilepickerClose,
  FilepickerHeading,
  FilepickerList,
  FilepickerListItem,
  FilepickerSeparator,
} from "../Filepicker";

import { GWForm, GWInput, GWInput2, GWSelect } from ".";
import type { Meta, StoryObj } from "@storybook/react";

import { Popover, PopoverContent, PopoverTrigger } from "../Popover";

import React, {
  ChangeEvent,
  ChangeEventHandler,
  Fragment,
  useState,
} from "react";

import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../Select";

import Button from "../Button";
import { ButtonBar } from "../ButtonBar";
import { Checkbox } from "../Checkbox";

import { GWCheckbox } from ".";

import Input from "../Input";
import { InputWithButton } from "../InputWithButton";

import { Textarea } from "../TextArea";
import { Separator } from "@radix-ui/react-dropdown-menu";
import { DropdownMenuSeparator } from "../Dropdown";

const meta = {
  title: "Components/GatewayForms",
  component: Filepicker,
} satisfies Meta<typeof Filepicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <Filepicker {...args}>content goes here...</Filepicker>
  ),
  argTypes: {},
};

export const BasicAuthFormPlugin = () => (
  <div className="flex flex-col p-2">
    <div className="flex flex-row py-2">
      <div className="flex items-center justify-center">
        <label
          htmlFor="add_header_username"
          className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Add username header:
        </label>
        <Checkbox id="add_header_username" className="m-2" />
      </div>
    </div>

    <div className="flex flex-row py-2">
      <div className="flex items-center justify-center">
        <label
          htmlFor="add_header_tags"
          className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Add tags header:
        </label>
        <Checkbox id="add_header_tags" className="m-2" />
      </div>
    </div>
    <div className="flex flex-row py-2">
      <div className="flex items-center justify-center">
        <label
          htmlFor="add_header_groups"
          className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Add groups header:
        </label>
        <Checkbox id="add_header_groups" className="m-2" />
      </div>
    </div>
  </div>
);

export const KeyAuthFormPlugin = () => (
  <div className="flex flex-col p-2">
    <div className="flex flex-row py-2">
      <div className="flex items-center justify-center">
        <label
          htmlFor="add_header_username"
          className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Add username header:
        </label>
        <Checkbox id="add_header_username" className="m-2" />
      </div>
    </div>

    <div className="flex flex-row py-2">
      <div className="flex items-center justify-center">
        <label
          htmlFor="add_header_tags"
          className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Add tags header:
        </label>
        <Checkbox id="add_header_tags" className="m-2" />
      </div>
    </div>
    <div className="flex flex-row py-2">
      <div className="flex items-center justify-center">
        <label
          htmlFor="add_header_groups"
          className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Add groups header:
        </label>
        <Checkbox id="add_header_groups" className="m-2" />
      </div>
    </div>

    <div className="flex flex-col py-2 sm:flex-row">
      <label
        htmlFor="add_key"
        className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
      >
        Key name:
      </label>
      <Input className="sm:w-max" id="add_key" placeholder="Insert name" />
    </div>
  </div>
);

export const ACLPlugin = () => {
  const [date, setDate] = React.useState<Date | undefined>(new Date());
  const [name, setName] = React.useState<string>(() => "Group 1");
  const [name2, setName2] = React.useState<string>(() => "Group 2");
  const [name3, setName3] = React.useState<string>(() => "No Entry");
  const [name4, setName4] = React.useState<string>(() => "Group 4");

  return (
    <div className="flex flex-col p-2">
      <div className="flex flex-row py-2">
        <div className="flex justify-center">
          <label
            htmlFor="add_variable"
            className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
          >
            Allow Groups:
          </label>
        </div>

        <div className="flex justify-start">
          <ButtonBar>
            <Input placeholder="Insert group name" value="Group 1" />
            <Button variant="outline" icon>
              <X />
            </Button>
          </ButtonBar>
        </div>
      </div>
      <div className="flex flex-row py-2">
        <div className="m-2 w-32"></div>
        <div className="flex justify-start">
          <ButtonBar>
            <Input placeholder="Insert group name" value="Group 2" />
            <Button variant="outline" icon>
              <X />
            </Button>
          </ButtonBar>
        </div>
      </div>

      <div className="flex flex-row py-2">
        <div className="m-2 w-32"></div>
        <div className="flex justify-start">
          <ButtonBar>
            <Button variant="outline" icon>
              <Plus />
            </Button>
          </ButtonBar>
        </div>
      </div>

      <div className="flex flex-row py-2">
        <div className="flex justify-center">
          <label
            htmlFor="add_variable"
            className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
          >
            Allow Tags:
          </label>
        </div>
        <div className="flex justify-start">
          <ButtonBar>
            <Input placeholder="No Entry" disabled />
            <Button disabled variant="outline" icon>
              <X />
            </Button>
          </ButtonBar>
        </div>
      </div>

      <div className="flex flex-row py-2">
        <div className="m-2 w-32"></div>
        <div className="flex justify-start">
          <ButtonBar>
            <Button variant="outline" icon>
              <Plus />
            </Button>
          </ButtonBar>
        </div>
      </div>

      <div className="flex flex-row py-2">
        <div className="flex justify-center">
          <label
            htmlFor="add_variable"
            className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
          >
            Deny Tags:
          </label>
        </div>
        <div className="flex justify-start">
          <ButtonBar>
            <Input placeholder="Insert group name" />
            <Button variant="outline" icon>
              <X />
            </Button>
          </ButtonBar>
        </div>
      </div>

      <div className="flex flex-row py-2">
        <div className="m-2 w-32"></div>
        <div className="flex justify-start">
          <ButtonBar>
            <Button variant="outline" icon>
              <Plus />
            </Button>
          </ButtonBar>
        </div>
      </div>
    </div>
  );
};

export const JSInboundPlugin = () => (
  <div className="flex flex-col p-2">
    <div className="flex flex-col py-2 sm:flex-row">
      <label
        htmlFor="add_script"
        className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
      >
        Script:
      </label>
      <Textarea id="add_script" placeholder="Insert Script" />
    </div>
  </div>
);

export const JSOutboundPlugin = () => (
  <div className="flex flex-col p-2">
    <div className="flex flex-col py-2 sm:flex-row">
      <label
        htmlFor="add_script"
        className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
      >
        Script:
      </label>

      <Textarea id="add_script" placeholder="No Entry" disabled />
    </div>
  </div>
);

export const RequestConverterPlugin = () => (
  <div className="flex flex-col p-2">
    <div className="flex flex-row py-2">
      <div className="flex items-center justify-center">
        <label
          htmlFor="omit_headers"
          className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Omit Headers:
        </label>
        <Checkbox id="omit_headers" className="m-2" />
      </div>
    </div>

    <div className="flex flex-row py-2">
      <div className="flex items-center justify-center">
        <label
          htmlFor="omit_queries"
          className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Omit Queries:
        </label>
        <Checkbox id="omit_queries" className="m-2" />
      </div>
    </div>
    <div className="flex flex-row py-2">
      <div className="flex items-center justify-center">
        <label
          htmlFor="omit_body"
          className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Omit Body:
        </label>
        <Checkbox id="omit_body" className="m-2" />
      </div>
    </div>
    <div className="flex flex-row py-2">
      <div className="flex items-center justify-center">
        <label
          htmlFor="omit_consumer"
          className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Omit Consumer:
        </label>
        <Checkbox id="omit_consumer" className="m-2" />
      </div>
    </div>
  </div>
);

export const InstantResponse = () => (
  <div className="flex flex-col p-2">
    <div className="flex flex-col py-2 sm:flex-row">
      <label
        htmlFor="omit_headers"
        className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
      >
        Status Code:
      </label>

      <Input
        className="sm:w-max"
        id="add_status_code"
        placeholder="200"
        value="200"
      />
    </div>
    <div className="flex flex-col py-2 sm:flex-row">
      <label
        htmlFor="omit_headers"
        className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
      >
        Content Type:
      </label>

      <Input className="sm:w-max" id="add_content_type" placeholder="/json" />
    </div>
    <div className="flex flex-col py-2 sm:flex-row">
      <label
        htmlFor="omit_headers"
        className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
      >
        Status Message:
      </label>

      <Textarea id="add_status_message" placeholder="Insert Text" />
    </div>
  </div>
);

export const NamespaceFileTarget = () => {
  const [date, setDate] = React.useState<Date | undefined>(new Date());

  const [namespace, setNamespace] = React.useState<string | undefined>(
    undefined
  );
  const [name2, setName2] = React.useState<string>(() => "Group 2");

  return (
    <div className="flex flex-col p-2">
      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="select_namespace"
          className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Namespace:
        </label>

        <Select onValueChange={setNamespace}>
          <SelectTrigger variant="primary">
            <SelectValue placeholder="Select a namespace" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="Example">Example</SelectItem>
              <SelectItem value="My-Namespace">My-Namespace</SelectItem>
              <SelectItem value="Namespace-with-a-very-long-name">
                Namespace-with-a-very-long-name
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>
      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="select_namespace"
          className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          File:
        </label>

        <Filepicker>
          {!namespace ? (
            <FilepickerHeading>
              Please select a namespace first!
            </FilepickerHeading>
          ) : (
            <div>
              <FilepickerHeading>{namespace}</FilepickerHeading>
              <FilepickerSeparator />
              <FilepickerHeading>
                <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
                  <BreadcrumbRoot>
                    <Breadcrumb noArrow>
                      <a href="#">Root-Folder</a>
                    </Breadcrumb>
                    <Breadcrumb>
                      <a href="#">Mid-Folder</a>
                    </Breadcrumb>
                    <Breadcrumb>
                      <a href="#">Sub-Folder</a>
                    </Breadcrumb>
                  </BreadcrumbRoot>
                </h3>
              </FilepickerHeading>
              <FilepickerSeparator />
              <FilepickerListItem icon={FolderUp}>..</FilepickerListItem>
              <FilepickerSeparator />
              <FilepickerList>
                {items.map((element) => (
                  <Fragment key={element.filename}>
                    <FilepickerClose>
                      <FilepickerListItem icon={element.icon}>
                        {element.filename}
                      </FilepickerListItem>
                    </FilepickerClose>
                    <FilepickerSeparator />
                  </Fragment>
                ))}
              </FilepickerList>
            </div>
          )}
        </Filepicker>
      </div>

      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="add_content_type"
          className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Content Type:
        </label>
        <Input
          className="sm:w-max"
          id="add_content_type"
          placeholder="e.g. image/jpg"
        />
      </div>
    </div>
  );
};

// Beginning of Mock Data

const Example_items: Listitem[] = [
  { filename: "image.jpg", icon: File },
  { filename: "image1.jpg", icon: SettingsIcon },
  { filename: "image2.jpg", icon: File },
  { filename: "image3.jpg", icon: MenuSquareIcon },
];

const MyNamespace_items: Listitem[] = [
  { filename: "hello.yaml", icon: File },
  { filename: "hello1.yaml", icon: File },
  { filename: "hello2.yaml", icon: File },
  { filename: "hello3.yaml", icon: File },
  { filename: "hello4.yaml", icon: File },
  { filename: "Readme.txt", icon: MenuSquareIcon },
  { filename: "Readme0.txt", icon: SettingsIcon },
  { filename: "Readme1.txt", icon: SettingsIcon },
];

// End of Mock Data

type Listitem = {
  filename: string;
  icon: LucideIcon;
};

const items: Listitem[] = [
  { filename: "image.jpg", icon: File },
  { filename: "image1.jpg", icon: SettingsIcon },
  { filename: "image2.jpg", icon: File },
  { filename: "image3.jpg", icon: MenuSquareIcon },
  { filename: "hello.yaml", icon: File },
  { filename: "hello1.yaml", icon: File },
  { filename: "hello2.yaml", icon: File },
  { filename: "hello3.yaml", icon: File },
  { filename: "hello4.yaml", icon: File },
  { filename: "Readme.txt", icon: MenuSquareIcon },
  { filename: "Readme0.txt", icon: SettingsIcon },
  { filename: "Readme1.txt", icon: SettingsIcon },
  { filename: "Readme2.txt", icon: SettingsIcon },
  { filename: "Readme3.txt", icon: SettingsIcon },
  { filename: "Readme4.txt", icon: File },
  { filename: "Readme5.txt", icon: File },
  { filename: "Readme6.txt", icon: MenuSquareIcon },
  { filename: "Readme7.txt", icon: SettingsIcon },
  { filename: "Readme8.txt", icon: File },
  { filename: "Readme9.txt", icon: SettingsIcon },
  { filename: "Readme10.txt", icon: File },
  { filename: "Readme11.txt", icon: File },
];

export const NamespaceVariableTarget = () => {
  const [namespace, setNamespace] = React.useState<string>(
    () => "My-Namespace"
  );
  return (
    <div className="flex flex-col p-2">
      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="select_namespace"
          className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Namespace:
        </label>

        <Select onValueChange={setNamespace}>
          <SelectTrigger variant="primary">
            <SelectValue placeholder="Select a namespace" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="Example">Example</SelectItem>
              <SelectItem value="My-Namespace">My-Namespace</SelectItem>
              <SelectItem value="Namespace-with-a-very-long-name">
                Namespace-with-a-very-long-name
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>
      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="add_variable"
          className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Variable:
        </label>

        <Input
          className="sm:w-max"
          id="add_variable"
          placeholder="insert name"
        />
      </div>

      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="add_content_type"
          className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Content Type:
        </label>

        <Input
          className="sm:w-max"
          id="add_content_type"
          placeholder="e.g. image/jpg"
        />
      </div>
    </div>
  );
};

export const WorkflowVariableTarget = () => {
  const [namespace, setNamespace] = React.useState<string>(
    () => "My-Namespace"
  );
  return (
    <div className="flex flex-col p-2">
      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="select_namespace"
          className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Namespace:
        </label>

        <Select onValueChange={setNamespace}>
          <SelectTrigger variant="primary">
            <SelectValue placeholder="Select a namespace" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="Example">Example</SelectItem>
              <SelectItem value="My-Namespace">My-Namespace</SelectItem>
              <SelectItem value="Namespace-with-a-very-long-name">
                Namespace-with-a-very-long-name
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>
      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="select_namespace"
          className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Workflow:
        </label>

        <Select onValueChange={setNamespace}>
          <SelectTrigger variant="primary">
            <SelectValue placeholder="Select a workflow" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="Workflow1">Workflow1</SelectItem>
              <SelectItem value="Workflow22">Workflow22</SelectItem>
              <SelectItem value="Workflow333">Workflow333</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>
      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="add_variable"
          className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Variable:
        </label>

        <Input
          className="sm:w-max"
          id="add_variable"
          placeholder="insert name"
        />
      </div>
      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="add_content_type"
          className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Content Type:
        </label>

        <Input
          className="sm:w-max"
          id="add_content_type"
          placeholder="e.g. image/jpg"
        />
      </div>
    </div>
  );
};

export const WorkflowTarget = () => {
  const [namespace, setNamespace] = React.useState<string>(
    () => "My-Namespace"
  );
  return (
    <div className="flex flex-col p-2">
      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="select_namespace"
          className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Namespace:
        </label>

        <Select onValueChange={setNamespace}>
          <SelectTrigger variant="primary">
            <SelectValue placeholder="Select a namespace" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="Example">Example</SelectItem>
              <SelectItem value="My-Namespace">My-Namespace</SelectItem>
              <SelectItem value="Namespace-with-a-very-long-name">
                Namespace-with-a-very-long-name
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>
      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="select_namespace"
          className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Workflow:
        </label>

        <Select onValueChange={setNamespace}>
          <SelectTrigger variant="primary">
            <SelectValue placeholder="Select a workflow" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="Workflow1">Workflow1</SelectItem>
              <SelectItem value="Workflow22">Workflow22</SelectItem>
              <SelectItem value="Workflow333">Workflow333</SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>
      <div className="flex flex-row py-2">
        <div className="flex items-center justify-center">
          <label
            htmlFor="asynchronous"
            className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
          >
            Asynchronous:
          </label>
        </div>
        <div className="flex items-center">
          <Checkbox id="asynchronous" />
        </div>
      </div>

      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="add_content_type"
          className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Content Type:
        </label>

        <Input
          className="sm:w-max"
          id="add_content_type"
          placeholder="e.g. image/jpg"
        />
      </div>
    </div>
  );
};

export const AllFormsDesign = () => {
  const [namespace, setNamespace] = React.useState<string>(
    () => "My-Namespace"
  );
  return (
    <div className="flex flex-col p-2">
      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="select_namespace"
          className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Namespace:
        </label>

        <Select onValueChange={setNamespace}>
          <SelectTrigger variant="primary">
            <SelectValue placeholder="Select a namespace" />
          </SelectTrigger>
          <SelectContent>
            <SelectGroup>
              <SelectItem value="Example">Example</SelectItem>
              <SelectItem value="My-Namespace">My-Namespace</SelectItem>
              <SelectItem value="Namespace-with-a-very-long-name">
                Namespace-with-a-very-long-name
              </SelectItem>
            </SelectGroup>
          </SelectContent>
        </Select>
      </div>
      <div className="flex flex-row py-2">
        <div className="flex items-center justify-center">
          <label
            htmlFor="asynchronous"
            className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
          >
            Asynchronous:
          </label>
        </div>
        <div className="flex items-center">
          <Checkbox id="asynchronous" />
        </div>
      </div>
      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="add_content_type"
          className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Content Type:
        </label>

        <Input
          className="sm:w-max"
          id="add_content_type"
          placeholder="e.g. image/jpg"
        />
      </div>
      <div className="flex flex-col py-2 sm:flex-row">
        <label
          htmlFor="omit_headers"
          className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Status Message:
        </label>

        <Textarea id="add_status_message" placeholder="Insert Text" />
      </div>
      <div className="flex flex-row py-2">
        <div className="flex justify-center">
          <label
            htmlFor="add_variable"
            className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
          >
            Allow Groups:
          </label>
        </div>

        <div className="flex justify-start">
          <ButtonBar>
            <Input placeholder="Insert group name" value="Group 1" />
            <Button variant="outline" icon>
              <X />
            </Button>
          </ButtonBar>
        </div>
      </div>
      <div className="flex flex-row py-2">
        <div className="m-2 w-32"></div>
        <div className="flex justify-start">
          <ButtonBar>
            <Input placeholder="Insert group name" value="Group 2" />
            <Button variant="outline" icon>
              <X />
            </Button>
          </ButtonBar>
        </div>
      </div>

      <div className="flex flex-row py-2">
        <div className="m-2 w-32"></div>
        <div className="flex justify-start">
          <ButtonBar>
            <Button variant="outline" icon>
              <Plus />
            </Button>
          </ButtonBar>
        </div>
      </div>
    </div>
  );
};

export const AllFormsFunctionalityDemo = () => {
  const [gwCheckbox, setgwCheckbox] = useState(false);
  const [namespace, setNamespace] = useState("init");
  const [value, setValue] = useState("");

  const handleChange = () => {
    setgwCheckbox(gwCheckbox ? false : true);
  };

  return (
    <div>
      <h3 className="font-bold">Data:</h3>
      <p>Checkbox: {gwCheckbox ? "TRUE" : "FALSE"}</p>
      <p>Select: {namespace}</p>
      <p>Input: {value}</p>
      <DropdownMenuSeparator />
      <GWSelect onValueChange={setNamespace}>Label</GWSelect>
      <GWCheckbox handleChange={handleChange} checked={gwCheckbox}>
        Asynchronous:
      </GWCheckbox>
      <GWInput2 onChange={setValue} value={value} placeholder="insert text...">
        Label2
      </GWInput2>
    </div>
  );
};

export const ShowForm = () => <GWForm></GWForm>;
