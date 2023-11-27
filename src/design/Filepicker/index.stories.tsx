import { File, Folder, FolderUp } from "lucide-react";

import {
  Filepicker,
  FilepickerHeading,
  FilepickerList,
  FilepickerListItem,
  FilepickerSeparator,
} from "./";

import type { Meta, StoryObj } from "@storybook/react";

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
  argTypes: {},
};

export const WithFewItems = () => (
  <Filepicker>
    <FilepickerHeading>Collection of Files</FilepickerHeading>
    <FilepickerSeparator />
    <div className="w-full hover:bg-gray-3 dark:hover:bg-gray-dark-3">
      <FilepickerListItem icon={FolderUp}>..</FilepickerListItem>
    </div>
    <FilepickerSeparator />
    <div className="w-full hover:bg-gray-3 dark:hover:bg-gray-dark-3">
      <FilepickerListItem icon={Folder}>Images</FilepickerListItem>
    </div>
    <FilepickerSeparator />
    <div className="w-full hover:bg-gray-3 dark:hover:bg-gray-dark-3">
      <FilepickerListItem icon={Folder}>Text</FilepickerListItem>
    </div>
    <FilepickerSeparator />
    <div className="w-full hover:bg-gray-3 dark:hover:bg-gray-dark-3">
      <FilepickerListItem icon={File}>Readme.txt</FilepickerListItem>
    </div>
    <FilepickerSeparator />
    <div className="w-full hover:bg-gray-3 dark:hover:bg-gray-dark-3">
      <FilepickerListItem asChild icon={File}>
        Icon.jpg
      </FilepickerListItem>
    </div>
  </Filepicker>
);

type Listitem = {
  filename: string;
  index: number;
};

const items: Listitem[] = [
  { filename: "image.jpg", index: 0 },
  { filename: "image1.jpg", index: 1 },
  { filename: "image2.jpg", index: 2 },
  { filename: "image3.jpg", index: 3 },
  { filename: "hello.yaml", index: 4 },
  { filename: "hello1.yaml", index: 5 },
  { filename: "hello2.yaml", index: 6 },
  { filename: "hello3.yaml", index: 7 },
  { filename: "hello4.yaml", index: 8 },
  { filename: "Readme.txt", index: 9 },
  { filename: "Readme0.txt", index: 10 },
  { filename: "Readme1.txt", index: 11 },
  { filename: "Readme2.txt", index: 12 },
  { filename: "Readme3.txt", index: 13 },
  { filename: "Readme4.txt", index: 14 },
  { filename: "Readme5.txt", index: 15 },
  { filename: "Readme6.txt", index: 16 },
  { filename: "Readme7.txt", index: 17 },
  { filename: "Readme8.txt", index: 18 },
  { filename: "Readme9.txt", index: 19 },
  { filename: "Readme10.txt", index: 20 },
  { filename: "Readme11.txt", index: 21 },
];

export const WithManyItems = () => (
  <Filepicker>
    <FilepickerHeading>Collection of Files</FilepickerHeading>
    <FilepickerSeparator />
    <div className="w-full hover:bg-gray-3 dark:hover:bg-gray-dark-3">
      <FilepickerListItem icon={FolderUp}>..</FilepickerListItem>
    </div>
    <FilepickerSeparator />
    <FilepickerList>
      {items.map((element) => (
        <div key={element.index}>
          <FilepickerListItem icon={File}>
            {element.filename}
          </FilepickerListItem>
          <FilepickerSeparator />
        </div>
      ))}
    </FilepickerList>
  </Filepicker>
);
