import * as DropdownMenuPrimitive from "@radix-ui/react-dropdown-menu";
import * as React from "react";

import { FC, PropsWithChildren } from "react";
import {
  Popover,
  PopoverClose,
  PopoverContent,
  PopoverTrigger,
} from "~/design/Popover";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "../Select";

import { Braces } from "lucide-react";
import Button from "../Button";
import { RxChevronDown } from "react-icons/rx";
import { twMergeClsx } from "~/util/helpers";

const VariablepickerSeparator = React.forwardRef<
  React.ElementRef<typeof DropdownMenuPrimitive.Separator>,
  React.ComponentPropsWithoutRef<typeof DropdownMenuPrimitive.Separator>
>(({ className, ...props }, ref) => (
  <DropdownMenuPrimitive.Separator
    ref={ref}
    className={twMergeClsx(
      "my-1 h-px bg-gray-3 dark:bg-gray-dark-3",
      className
    )}
    {...props}
  />
));

VariablepickerSeparator.displayName =
  DropdownMenuPrimitive.Separator.displayName;

type VariablepickerPropsType = PropsWithChildren & {
  buttonText: string;
  value?: any;
  onChange?: (variable: any) => void;
  onValueChange?: (value: any) => void;
};

const Item = React.forwardRef<
  React.ElementRef<typeof SelectItem>,
  React.ComponentPropsWithoutRef<typeof SelectItem>
>(({ className, ...props }, ref) => <SelectItem {...props} />);

Item.displayName = SelectItem.displayName;

const Variablepicker: FC<VariablepickerPropsType> = ({
  children,
  buttonText,
  value,
  onChange,
  onValueChange,
}) => (
  <Select value={value} onValueChange={onValueChange}>
    <SelectTrigger className="w-52" variant="outline">
      <SelectValue placeholder={buttonText}>{buttonText}</SelectValue>
    </SelectTrigger>
    <SelectContent onChange={onChange}>
      <SelectGroup>{children}</SelectGroup>
    </SelectContent>
  </Select>
);

type VariablepickerItemProps = PropsWithChildren & {
  props?: object;
  value: any;
  disabled?: boolean;
};

/*
type VariablepickerItemProps = PropsWithChildren & {
  props?: object;
  onChange: React.FormEventHandler;
  value: {
    name: string;
    checksum: string;
    createdAt: string;
    updatedAt: string;
    size: string;
    mimeType: string;
  };
};
*/

const VariablepickerHeading: FC<PropsWithChildren> = ({ children }) => (
  <div className="px-2 text-sm font-semibold text-gray-9 dark:text-gray-dark-9">
    <div className="flex items-center px-2">
      <div className="w-max">
        <Braces className="h-4 w-4" aria-hidden="true" />
      </div>
      <div className="whitespace-nowrap px-3 py-2 text-sm">{children}</div>
    </div>
  </div>
);

const VariablepickerError: FC<VariablepickerPropsType> = ({
  children,
  buttonText,
}) => (
  <Popover modal>
    <PopoverTrigger asChild>
      <Button variant="outline" className="w-52">
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

const VariablepickerMessage2: FC<PropsWithChildren> = ({ children }) => (
  <SelectLabel className="p-2 text-sm text-gray-9 dark:text-gray-dark-9">
    {children}
  </SelectLabel>
);

/*
Vorher:

const VariablepickerHeading: FC<PropsWithChildren> = ({ children }) => (
  <div className="px-2 text-sm font-semibold text-gray-9 dark:text-gray-dark-9">
    {children}
  </div>
);
*/

/*
  <div className="flex items-center px-2">
  <div className="w-max">
    <Icon className="h-4 w-4" aria-hidden="true" />
  </div>
  <div className="whitespace-nowrap px-3 py-2 text-sm">{children}</div>
</div>

INSIDE:
<Braces className="h-5" /> Variables in {path}


*/
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
  VariablepickerError,
  VariablepickerHeading,
  VariablepickerItem,
  VariablepickerSeparator,
  VariablepickerMessage,
  VariablepickerMessage2,
};
