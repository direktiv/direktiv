import * as React from "react";

import { FC, PropsWithChildren } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../Select";

import Button from "../Button";
import { DropdownMenuSeparator } from "../Dropdown";
import { FileJson } from "lucide-react";
import { RxChevronDown } from "react-icons/rx";

const VariablepickerSeparator = DropdownMenuSeparator;

type VariablepickerCloseProps = PropsWithChildren & {
  onClick: React.MouseEventHandler;
};

const VariablepickerClose: FC<VariablepickerCloseProps> = ({
  children,
  onClick,
}) => <Button onClick={onClick}>{children}</Button>;

type VariablepickerProps = PropsWithChildren & {
  buttonText: string;
  value?: string;
  onValueChange?: (value: string) => void;
};

const Variablepicker: FC<VariablepickerProps> = ({
  children,
  buttonText,
  value,
  onValueChange,
}) => (
  <Select value={value} onValueChange={onValueChange}>
    <SelectTrigger className="w-64" variant="outline">
      <SelectValue placeholder={buttonText}>{buttonText}</SelectValue>
    </SelectTrigger>
    <SelectContent>
      <SelectGroup>{children}</SelectGroup>
    </SelectContent>
  </Select>
);

const VariablepickerHeading: FC<PropsWithChildren> = ({ children }) => (
  <div className="px-2 text-sm font-semibold text-gray-9 dark:text-gray-dark-9">
    <div className="flex items-center px-2">
      <div className="w-max">
        <FileJson className="h-4 w-4" aria-hidden="true" />
      </div>
      <div className="whitespace-nowrap px-3 py-2 text-sm">{children}</div>
    </div>
  </div>
);

const VariablepickerError: FC<VariablepickerProps> = ({
  children,
  buttonText,
}) => (
  <Popover modal>
    <PopoverTrigger asChild>
      <Button variant="outline" className="w-64">
        <span>{buttonText}</span>
        <RxChevronDown />
      </Button>
    </PopoverTrigger>
    <PopoverContent
      className="w-screen min-w-full bg-gray-1 dark:bg-gray-dark-1 lg:w-3/4"
      align="start"
    >
      {children}
    </PopoverContent>
  </Popover>
);
const VariablepickerMessage: FC<PropsWithChildren> = ({ children }) => (
  <div className="p-2 text-sm text-gray-9 dark:text-gray-dark-9">
    {children}
  </div>
);

type VariablepickerItemProps = PropsWithChildren & {
  props?: object;
  value: string;
  disabled?: boolean;
};

const VariablepickerItem: FC<VariablepickerItemProps> = ({
  props,
  value,
  children,
  disabled,
}) => (
  <SelectItem disabled={disabled} value={value} {...props}>
    {children}
  </SelectItem>
);

export {
  Variablepicker,
  VariablepickerClose,
  VariablepickerError,
  VariablepickerHeading,
  VariablepickerItem,
  VariablepickerSeparator,
  VariablepickerMessage,
};
