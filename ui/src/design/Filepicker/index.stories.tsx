import { Breadcrumb, BreadcrumbRoot } from "../Breadcrumbs";

import { Dialog, DialogContent, DialogTrigger } from "../Dialog";
import {
  File,
  Folder,
  FolderUp,
  LucideIcon,
  MenuSquareIcon,
  SettingsIcon,
} from "lucide-react";

import {
  Filepicker,
  FilepickerClose,
  FilepickerHeading,
  FilepickerList,
  FilepickerListItem,
} from "./";
import { Fragment, useState } from "react";

import type { Meta, StoryObj } from "@storybook/react";

import Button from "../Button";
import { ButtonBar } from "../ButtonBar";

import Input from "../Input";
import { DropdownMenuSeparator } from "../Dropdown";

const meta = {
  title: "Components/Filepicker",
  component: Filepicker,
} satisfies Meta<typeof Filepicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <Filepicker {...args}>content goes here...</Filepicker>
  ),
  args: {
    buttonText: "Browse Files",
  },
  argTypes: {},
};

export const WithFewItems = () => (
  <Filepicker buttonText="Browse Files">
    <FilepickerHeading>Collection of Files</FilepickerHeading>
    <DropdownMenuSeparator />
    <FilepickerListItem icon={FolderUp}>..</FilepickerListItem>
    <DropdownMenuSeparator />
    <FilepickerListItem icon={Folder}>Images</FilepickerListItem>
    <DropdownMenuSeparator />
    <FilepickerListItem icon={Folder}>Text</FilepickerListItem>
    <DropdownMenuSeparator />
    <FilepickerListItem icon={File}>Readme.txt</FilepickerListItem>
    <DropdownMenuSeparator />
    <FilepickerListItem icon={File}>Icon.jpg</FilepickerListItem>
  </Filepicker>
);

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

export const WithManyItemsBreadcrumbHeadingAndCloseFunctionAtItemClick = () => {
  const [inputValue, setInputValue] = useState("");

  return (
    <ButtonBar>
      <Filepicker buttonText="Browse Files" className="w-96">
        <FilepickerHeading>
          <BreadcrumbRoot className="py-3">
            <Breadcrumb noArrow>
              <a href="#">My-namespace</a>
            </Breadcrumb>
            <Breadcrumb className="h-5 hover:underline">
              <a href="#">My-folder</a>
            </Breadcrumb>
            <Breadcrumb className="h-5 hover:underline">
              <a href="#">My-subfolder</a>
            </Breadcrumb>
          </BreadcrumbRoot>
        </FilepickerHeading>
        <DropdownMenuSeparator />
        <div className="h-auto w-full cursor-pointer p-0 font-normal text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent">
          <FilepickerListItem icon={FolderUp}>..</FilepickerListItem>
        </div>
        <DropdownMenuSeparator />

        <FilepickerList>
          {items.map((element) => (
            <Fragment key={element.filename}>
              <FilepickerClose
                className="h-auto w-full text-gray-11 opacity-70 hover:underline dark:text-gray-dark-11"
                onClick={() => {
                  setInputValue(element.filename);
                }}
              >
                <FilepickerListItem icon={element.icon}>
                  {element.filename}
                </FilepickerListItem>
              </FilepickerClose>

              <DropdownMenuSeparator />
            </Fragment>
          ))}
        </FilepickerList>
      </Filepicker>
      <Input
        placeholder="No File selected"
        value={inputValue}
        onChange={(e) => {
          setInputValue(e.target.value);
        }}
      />
    </ButtonBar>
  );
};

export const InAModal = () => {
  const [inputValue, setInputValue] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>
        <Button>Open Dialog Menu</Button>
      </DialogTrigger>
      <DialogContent>
        <ButtonBar>
          <Filepicker buttonText="Open Filepicker Menu" className="w-96">
            <FilepickerHeading>
              <BreadcrumbRoot className="py-3">
                <Breadcrumb noArrow>
                  <a href="#">My-namespace</a>
                </Breadcrumb>
                <Breadcrumb className="h-5 hover:underline">
                  <a href="#">My-folder</a>
                </Breadcrumb>
                <Breadcrumb className="h-5 hover:underline">
                  <a href="#">My-subfolder</a>
                </Breadcrumb>
              </BreadcrumbRoot>
            </FilepickerHeading>
            <DropdownMenuSeparator />
            <div className="h-auto w-full cursor-pointer p-0 font-normal text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent">
              <FilepickerListItem icon={FolderUp}>..</FilepickerListItem>
            </div>
            <DropdownMenuSeparator />

            <FilepickerList>
              {items.map((element) => (
                <Fragment key={element.filename}>
                  <FilepickerClose
                    className="h-auto w-full text-gray-11 opacity-70 hover:underline dark:text-gray-dark-11"
                    onClick={() => {
                      setInputValue(element.filename);
                    }}
                  >
                    <FilepickerListItem icon={element.icon}>
                      {element.filename}
                    </FilepickerListItem>
                  </FilepickerClose>

                  <DropdownMenuSeparator />
                </Fragment>
              ))}
            </FilepickerList>
          </Filepicker>
          <Input
            placeholder="No File selected"
            value={inputValue}
            onChange={(e) => {
              setInputValue(e.target.value);
            }}
          />
        </ButtonBar>
      </DialogContent>
    </Dialog>
  );
};
