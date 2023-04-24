import type { Meta, StoryObj } from "@storybook/react";
import { RadioGroup, RadioGroupItem } from "../RadioGroup";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "./index";
import Button from "../Button";
import Input from "../Input";
import { useState } from "react";

const meta = {
  title: "Components/Sheet",
  component: Sheet,
} satisfies Meta<typeof Sheet>;

export default meta;
type Story = StoryObj<typeof meta>;

const SHEET_SIZES = ["sm", "default", "lg", "xl", "full", "content"] as const;

type SheetSize = (typeof SHEET_SIZES)[number];

const SHEET_DICRECTIONS = ["left", "right", "top", "bottom"] as const;
type SheetDirection = (typeof SHEET_DICRECTIONS)[number];

const StoryCompontnt = () => {
  const [size, setSize] = useState<SheetSize>("default");
  const [direction, setDirection] = useState<SheetDirection>("top");
  return (
    <Sheet>
      <SheetTrigger asChild>
        <Button>Open {size} sheet</Button>
      </SheetTrigger>
      <SheetContent position={direction} size={size}>
        <SheetHeader>
          <SheetTitle>Edit profile</SheetTitle>
          <SheetDescription>
            {`Make changes to your profile here. Click save when you're done.`}
          </SheetDescription>
        </SheetHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <label
              htmlFor="name"
              className="text-right text-black dark:text-white"
            >
              Name
            </label>
            <Input id="name" value="Pedro Duarte" className="col-span-3" />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <label
              htmlFor="username"
              className="text-right text-black dark:text-white"
            >
              Username
            </label>
            <Input id="username" value="@peduarte" className="col-span-3" />
          </div>
        </div>
        <SheetFooter>
          <Button type="submit">Save changes</Button>
        </SheetFooter>
      </SheetContent>
      <div className="flex flex-col space-y-8">
        <RadioGroup
          defaultValue={size}
          onValueChange={(value) => setSize(value as SheetSize)}
        >
          <div className="grid grid-cols-2 gap-2">
            {SHEET_SIZES.map((size, index) => (
              <div
                key={`${size}-${index}`}
                className="flex items-center space-x-2"
              >
                <RadioGroupItem value={size} id={size} />
                <label htmlFor={size}>{size}</label>
              </div>
            ))}
          </div>
        </RadioGroup>
        <RadioGroup
          defaultValue={direction}
          onValueChange={(value) => setDirection(value as SheetDirection)}
        >
          <div className="grid grid-cols-2 gap-2">
            {SHEET_DICRECTIONS.map((direction, index) => (
              <div
                key={`${direction}-${index}`}
                className="flex items-center space-x-2"
              >
                <RadioGroupItem value={direction} id={direction} />
                <label
                  htmlFor={direction}
                  className="text-black dark:text-white"
                >
                  {direction}
                </label>
              </div>
            ))}
          </div>
        </RadioGroup>
      </div>
    </Sheet>
  );
};
export const Default: Story = {
  render: () => <StoryCompontnt />,
};

export const DefaultOpen = () => {
  const [size, setSize] = useState<SheetSize>("default");
  return (
    <Sheet defaultOpen>
      <SheetTrigger asChild>
        <Button>Open {size} sheet</Button>
      </SheetTrigger>
      <SheetContent position="left" size={size}>
        <SheetHeader>
          <SheetTitle>Edit profile</SheetTitle>
          <SheetDescription>
            {`Make changes to your profile here. Click save when you're done.`}
          </SheetDescription>
        </SheetHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <label
              htmlFor="name"
              className="text-right  text-black dark:text-white"
            >
              Name
            </label>
            <Input id="name" value="Pedro Duarte" className="col-span-3" />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <label
              htmlFor="username"
              className="text-right text-black dark:text-white"
            >
              Username
            </label>
            <Input id="username" value="@peduarte" className="col-span-3" />
          </div>
        </div>
        <SheetFooter>
          <Button type="submit">Save changes</Button>
        </SheetFooter>
      </SheetContent>
      <RadioGroup
        defaultValue={size}
        onValueChange={(value) => setSize(value as SheetSize)}
      >
        <div className="grid grid-cols-2 gap-2">
          {SHEET_SIZES.map((size, index) => (
            <div
              key={`${size}-${index}`}
              className="flex items-center space-x-2"
            >
              <RadioGroupItem value={size} id={size} />
              <label htmlFor={size}>{size}</label>
            </div>
          ))}
        </div>
      </RadioGroup>
    </Sheet>
  );
};
