import { Breadcrumb, BreadcrumbRoot } from "../Breadcrumbs";

import {
  CheckCircle2,
  File,
  Folder,
  FolderOpen,
  FolderUp,
  Home,
  LucideIcon,
  MenuSquareIcon,
  SettingsIcon,
} from "lucide-react";

import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogTitle,
  DialogTrigger,
} from "../Dialog";

import {
  Filepicker,
  FilepickerClose,
  FilepickerHeading,
  FilepickerList,
  FilepickerListItem,
  FilepickerSelectButton,
  FilepickerSeparator,
} from "./";
import { Fragment, useState } from "react";

import type { Meta, StoryObj } from "@storybook/react";

import Button from "../Button";
import { ButtonBar } from "../ButtonBar";

import Input from "../Input";

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
        <FilepickerSeparator />
        <div className="h-auto w-full cursor-pointer p-0 font-normal text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent">
          <FilepickerListItem icon={FolderUp}>..</FilepickerListItem>
        </div>
        <FilepickerSeparator />

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

              <FilepickerSeparator />
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
            <FilepickerSeparator />
            <div className="h-auto w-full cursor-pointer p-0 font-normal text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent">
              <FilepickerListItem icon={FolderUp}>..</FilepickerListItem>
            </div>
            <FilepickerSeparator />

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

                  <FilepickerSeparator />
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

type FolderListItem = {
  filename: string;
  icon: LucideIcon;
  level: number;
};

const itemsLevel1: FolderListItem[] = [
  { filename: "projectA", icon: Folder, level: 1 },
  { filename: "projectB", icon: Folder, level: 1 },
  { filename: "projectC", icon: Folder, level: 1 },
];

const itemslevel2: FolderListItem[] = [
  { filename: "scripts", icon: Folder, level: 2 },
  { filename: "src", icon: Folder, level: 2 },
  { filename: "assets", icon: Folder, level: 2 },
  { filename: "locales", icon: Folder, level: 2 },
  { filename: "test", icon: Folder, level: 2 },
];

const itemslevel3: FolderListItem[] = [
  { filename: "mockup", icon: File, level: 3 },
  { filename: "image", icon: File, level: 3 },
  { filename: "text", icon: File, level: 3 },
  { filename: "music", icon: File, level: 3 },
];

export const InAModalWithFolders = () => {
  const [inputValue, setInputValue] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [elements, setElements] = useState(itemsLevel1);
  const [breadcrumb, setBreadcrumb] = useState("");
  const [secondBreadcrumb, setSecondBreadcrumb] = useState("");
  const [topLevel, setTopLevel] = useState(true);
  const [firstLevel, setFirstLevel] = useState(false);
  const [secondLevel, setSecondLevel] = useState(false);

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>
        <Button>Open Dialog Menu</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogTitle>
          <FolderOpen />
          Relocate
        </DialogTitle>
        <ButtonBar>
          <Filepicker buttonText="Select Folder" className="w-96">
            <FilepickerHeading>
              <BreadcrumbRoot className="py-3">
                <Breadcrumb
                  noArrow
                  onClick={() => {
                    setTopLevel(true);
                    setFirstLevel(false);
                    setSecondLevel(false);
                    setElements(itemsLevel1);
                  }}
                  className="h-5 hover:underline"
                >
                  <Home />
                </Breadcrumb>
                {firstLevel && (
                  <Breadcrumb className="h-5 hover:underline">
                    {breadcrumb}
                  </Breadcrumb>
                )}
                {secondLevel && (
                  <Fragment>
                    <Breadcrumb className="h-5 hover:underline">
                      {breadcrumb}
                    </Breadcrumb>
                    <Breadcrumb className="h-5 hover:underline">
                      {secondBreadcrumb}
                    </Breadcrumb>
                  </Fragment>
                )}
              </BreadcrumbRoot>
            </FilepickerHeading>
            <FilepickerSeparator />
            {!topLevel && (
              <Fragment>
                <div
                  className="h-auto w-full cursor-pointer p-0 font-normal text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent"
                  onClick={() => {
                    if (firstLevel) {
                      setSecondLevel(false);
                      setFirstLevel(false);
                      setTopLevel(true);
                      setElements(itemsLevel1);
                    }
                    if (secondLevel) {
                      setSecondLevel(false);
                      setFirstLevel(true);
                      setTopLevel(false);
                      setElements(itemslevel2);
                    }
                  }}
                >
                  <FilepickerListItem icon={FolderUp}>..</FilepickerListItem>
                </div>
                <FilepickerSeparator />
              </Fragment>
            )}

            <FilepickerList>
              {elements.map((element) => (
                <Fragment key={element.filename}>
                  {secondLevel && (
                    <div className="cursor-not-allowed text-gray-11 opacity-70 hover:bg-gray-3  focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:hover:bg-gray-dark-3 dark:focus:bg-transparent">
                      <FilepickerListItem icon={element.icon}>
                        {element.filename}
                      </FilepickerListItem>
                    </div>
                  )}
                  {!secondLevel && (
                    <div className="group flex h-auto w-full cursor-pointer items-center justify-between hover:bg-gray-3 dark:hover:bg-gray-dark-3">
                      <div
                        onClick={() => {
                          if (element.level === 1) {
                            setElements(itemslevel2);
                            setTopLevel(false);
                            setFirstLevel(true);
                            setSecondLevel(false);
                            setBreadcrumb(element.filename);
                          }
                          if (element.level === 2) {
                            setElements(itemslevel3);
                            setTopLevel(false);
                            setFirstLevel(false);
                            setSecondLevel(true);
                            setSecondBreadcrumb(element.filename);
                          }
                        }}
                        className="text-gray-11 hover:bg-gray-3 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:hover:bg-gray-dark-3 dark:focus:bg-transparent"
                      >
                        <FilepickerListItem icon={element.icon}>
                          {element.filename}
                        </FilepickerListItem>
                      </div>

                      <div className="h-auto px-4 py-2 opacity-0 group-hover:opacity-100">
                        <FilepickerClose
                          onClick={() => {
                            if (element.level === 1) {
                              setInputValue(element.filename);
                            } else {
                              if (!breadcrumb) {
                                setInputValue(element.filename);
                              }
                              setInputValue(
                                breadcrumb + "/" + element.filename
                              );
                            }
                          }}
                        >
                          <Button className="" size="sm">
                            <CheckCircle2 />
                            Select
                          </Button>
                        </FilepickerClose>
                      </div>
                    </div>
                  )}
                </Fragment>
              ))}
            </FilepickerList>
          </Filepicker>
          <Input
            placeholder="No Folder selected"
            value={inputValue}
            onChange={(e) => {
              setInputValue(e.target.value);
            }}
          />
        </ButtonBar>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">Cancel</Button>
          </DialogClose>
          <Button data-testid="node-rename-submit" type="submit">
            <FolderOpen />
            Relocate
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export const InAModalWithFolders2 = () => {
  const [inputValue, setInputValue] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [elements, setElements] = useState(itemsLevel1);
  const [breadcrumb, setBreadcrumb] = useState("");
  const [secondBreadcrumb, setSecondBreadcrumb] = useState("");
  const [topLevel, setTopLevel] = useState(true);
  const [firstLevel, setFirstLevel] = useState(false);
  const [secondLevel, setSecondLevel] = useState(false);

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DialogTrigger asChild>
        <Button>Open Dialog Menu</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogTitle>
          <FolderOpen />
          Relocate
        </DialogTitle>
        <ButtonBar>
          <Filepicker buttonText="Select Folder" className="w-96">
            <FilepickerHeading>
              <BreadcrumbRoot className="py-3">
                <Breadcrumb
                  noArrow
                  onClick={() => {
                    setTopLevel(true);
                    setFirstLevel(false);
                    setSecondLevel(false);
                    setElements(itemsLevel1);
                  }}
                  className="h-5 hover:underline"
                >
                  <Home />
                </Breadcrumb>
                {firstLevel && (
                  <Breadcrumb className="h-5 hover:underline">
                    {breadcrumb}
                  </Breadcrumb>
                )}
                {secondLevel && (
                  <Fragment>
                    <Breadcrumb className="h-5 hover:underline">
                      {breadcrumb}
                    </Breadcrumb>
                    <Breadcrumb className="h-5 hover:underline">
                      {secondBreadcrumb}
                    </Breadcrumb>
                  </Fragment>
                )}
              </BreadcrumbRoot>
            </FilepickerHeading>
            <FilepickerSeparator />
            {!topLevel && (
              <Fragment>
                <div
                  className="h-auto w-full cursor-pointer p-0 font-normal text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent"
                  onClick={() => {
                    if (firstLevel) {
                      setSecondLevel(false);
                      setFirstLevel(false);
                      setTopLevel(true);
                      setElements(itemsLevel1);
                    }
                    if (secondLevel) {
                      setSecondLevel(false);
                      setFirstLevel(true);
                      setTopLevel(false);
                      setElements(itemslevel2);
                    }
                  }}
                >
                  <FilepickerListItem icon={FolderUp}>..</FilepickerListItem>
                </div>
                <FilepickerSeparator />
              </Fragment>
            )}

            <FilepickerList>
              {elements.map((element) => (
                <Fragment key={element.filename}>
                  {secondLevel && (
                    <div className="cursor-not-allowed text-gray-11 opacity-70 hover:bg-gray-3  focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:hover:bg-gray-dark-3 dark:focus:bg-transparent">
                      <FilepickerListItem icon={element.icon}>
                        {element.filename}
                      </FilepickerListItem>
                    </div>
                  )}
                  {!secondLevel && (
                    <div className="group flex h-auto w-full cursor-pointer items-center justify-between hover:bg-gray-3 dark:hover:bg-gray-dark-3">
                      <div
                        onClick={() => {
                          if (element.level === 1) {
                            setElements(itemslevel2);
                            setTopLevel(false);
                            setFirstLevel(true);
                            setSecondLevel(false);
                            setBreadcrumb(element.filename);
                          }
                          if (element.level === 2) {
                            setElements(itemslevel3);
                            setTopLevel(false);
                            setFirstLevel(false);
                            setSecondLevel(true);
                            setSecondBreadcrumb(element.filename);
                          }
                        }}
                        className="text-gray-11 hover:bg-gray-3 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:hover:bg-gray-dark-3 dark:focus:bg-transparent"
                      >
                        <FilepickerListItem icon={element.icon}>
                          {element.filename}
                        </FilepickerListItem>
                      </div>

                      <div className="h-auto px-4 py-2 opacity-0 group-hover:opacity-100">
                        <FilepickerSelectButton
                          onClick={() => {
                            if (element.level === 1) {
                              setInputValue(element.filename);
                            } else {
                              if (!breadcrumb) {
                                setInputValue(element.filename);
                              }
                              setInputValue(
                                breadcrumb + "/" + element.filename
                              );
                            }
                          }}
                        >
                          Select
                        </FilepickerSelectButton>
                      </div>
                    </div>
                  )}
                </Fragment>
              ))}
            </FilepickerList>
          </Filepicker>
          <Input
            placeholder="No Folder selected"
            value={inputValue}
            onChange={(e) => {
              setInputValue(e.target.value);
            }}
          />
        </ButtonBar>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">Cancel</Button>
          </DialogClose>
          <Button data-testid="node-rename-submit" type="submit">
            <FolderOpen />
            Relocate
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
