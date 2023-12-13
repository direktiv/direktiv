import { Breadcrumb, BreadcrumbRoot } from "../Breadcrumbs";

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
  FilepickerSeparator,
} from "./";

import type { Meta, StoryObj } from "@storybook/react";
import { Fragment } from "react";

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
    <FilepickerSeparator />
    <FilepickerListItem icon={FolderUp}>..</FilepickerListItem>
    <FilepickerSeparator />
    <FilepickerListItem icon={Folder}>Images</FilepickerListItem>
    <FilepickerSeparator />
    <FilepickerListItem icon={Folder}>Text</FilepickerListItem>
    <FilepickerSeparator />
    <FilepickerListItem icon={File}>Readme.txt</FilepickerListItem>
    <FilepickerSeparator />
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

export const WithManyItemsBreadcrumbHeadingAndCloseFunctionAtItemClick = () => (
  <Filepicker buttonText="Browse Files">
    <FilepickerHeading>
      <BreadcrumbRoot>
        <Breadcrumb noArrow>
          <a href="#">My-namespace</a>
        </Breadcrumb>
        <Breadcrumb>
          <a href="#">My-folder</a>
        </Breadcrumb>
        <Breadcrumb>
          <a href="#">My-subfolder</a>
        </Breadcrumb>
      </BreadcrumbRoot>
    </FilepickerHeading>
    <FilepickerSeparator />
    <FilepickerListItem icon={FolderUp}>..</FilepickerListItem>
    <FilepickerSeparator />
    <FilepickerList>
      {items.map((element) => (
        <Fragment key={element.filename}>
          <FilepickerListItem icon={element.icon} asChild>
            <FilepickerClose>{element.filename}</FilepickerClose>
          </FilepickerListItem>
          <FilepickerSeparator />
        </Fragment>
      ))}
    </FilepickerList>
  </Filepicker>
);
