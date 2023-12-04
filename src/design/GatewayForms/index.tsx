import {
  ChangeEventHandler,
  FC,
  FormEventHandler,
  PropsWithChildren,
} from "react";
import { Fragment, useState } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../Select";

import Button from "~/design/Button";
import { Checkbox } from "../Checkbox";
import Input from "../Input";
import { twMergeClsx } from "~/util/helpers";

type FilepickerPropsType = PropsWithChildren & {
  className?: string;
};

const Filepicker: FC<FilepickerPropsType> = ({ className, children }) => (
  <div className={twMergeClsx("", className)}>
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="primary" data-testid="filepicker-button">
          <div className="relative">Browse Files</div>
        </Button>
      </PopoverTrigger>
      <PopoverContent className="bg-gray-1 dark:bg-gray-dark-1" align="start">
        {children}
      </PopoverContent>
    </Popover>
  </div>
);

type CheckboxPropsType = PropsWithChildren & {
  className?: string;
  checked?: boolean;
  handleChange?: FormEventHandler;
};

const GWCheckbox: FC<CheckboxPropsType> = ({
  className,
  children,
  checked,
  handleChange,
}) => (
  <div className="flex flex-row py-2">
    <div className="flex items-center justify-center">
      <label
        htmlFor="GWCheckbox"
        className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
      >
        {children}
      </label>
    </div>
    <div className="flex items-center">
      <Checkbox onClick={handleChange} checked={checked} id="GWCheckbox" />
    </div>
  </div>
);

const GWForm: FC = () => {
  const [gwCheckbox, setgwCheckbox] = useState<boolean>(() => true);

  const handleChange = () => {
    setgwCheckbox(gwCheckbox ? false : true);
  };

  return (
    <div>
      <GWCheckbox handleChange={handleChange} checked={gwCheckbox}>
        Asynchronous:
      </GWCheckbox>
    </div>
  );
};

type SelectPropsType = PropsWithChildren & {
  className?: string;
  onValueChange?: React.Dispatch<React.SetStateAction<string>>;
};

const GWSelect: FC<SelectPropsType> = ({
  className,
  children,
  onValueChange,
}) => (
  <div className="flex flex-col py-2 sm:flex-row">
    <label
      htmlFor="select_namespace"
      className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
    >
      {children}
    </label>
    <Select onValueChange={onValueChange}>
      <SelectTrigger variant="primary">
        <SelectValue placeholder="Select a namespace" />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          {items.map((element) => (
            <Fragment key={element.name}>
              <SelectItem value={element.name}>{element.name}</SelectItem>
            </Fragment>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  </div>
);

type Namespace = {
  name: string;
};

const items: Namespace[] = [
  { name: "Example" },
  { name: "My-Namespace" },
  { name: "Namespace-with-a-very-long-name" },
];

type InputPropsType = PropsWithChildren & {
  className?: string;
  onValueChange?: ChangeEventHandler;
  // onValueChange?: React.Dispatch<React.SetStateAction<string>>;
  value?: string;
  placeholder: string;
};

const GWInput: FC<InputPropsType> = ({
  className,
  children,
  onValueChange,
  value,
  placeholder,
}) => (
  <div className="flex flex-col py-2 sm:flex-row">
    <label
      htmlFor="add_key"
      className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
    >
      {children}
    </label>
    <Input
      onChange={onValueChange}
      className="sm:w-max"
      id="add_key"
      placeholder={placeholder}
      value={value}
    />
  </div>
);

export { Filepicker, GWCheckbox, GWForm, GWInput, GWSelect };
