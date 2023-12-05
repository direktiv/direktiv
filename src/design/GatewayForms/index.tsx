import {
  ChangeEvent,
  ChangeEventHandler,
  FC,
  FormEventHandler,
  MouseEventHandler,
  PropsWithChildren,
} from "react";
import { Fragment, useState } from "react";
import { Plus, X } from "lucide-react";
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
import { ButtonBar } from "../ButtonBar";
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
      <GWInputButtonList>Test </GWInputButtonList>
      <GWInputButton placeholder="Insert Group Name">Old Input</GWInputButton>
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

type InputPropsType2 = PropsWithChildren & {
  className?: string;
  onValueChange?: ChangeEventHandler;
  value?: string;
  placeholder?: string;
  handleChange?: any;
  onChange?: (newValue: string) => void;
};

const GWInput2: FC<InputPropsType2> = ({
  value,
  children,
  placeholder,
  onChange,
}) => {
  const x = 0;

  return (
    <div className="flex flex-col py-2 sm:flex-row">
      <label
        htmlFor="add_key"
        className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
      >
        {children}
      </label>
      <Input
        onChange={(e) => {
          onChange?.(e.target.value);
        }}
        className="sm:w-max"
        id="add_key"
        placeholder={placeholder}
        value={value}
      />
    </div>
  );
};

type InputButtonProps = PropsWithChildren & {
  className?: string;
  onValueChange?: ChangeEventHandler;
  value?: string;
  placeholder?: string;
  onChange?: (event: React.ChangeEvent<HTMLInputElement>) => void; // TODO: change to the same type as in GWInput2
  rowFilled?: boolean;
  onClick?: MouseEventHandler;
};

const GWInputButton: FC<InputButtonProps> = ({
  value,
  children,
  placeholder,
  onChange,
  onClick,
  rowFilled,
}) => {
  const b = 0;

  return (
    <div className="flex flex-col p-2">
      <div className="flex flex-row py-2">
        <div className="flex justify-center">
          <label
            htmlFor="add_variable"
            className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
          >
            {children}
          </label>
        </div>
        <div className="flex justify-start">
          <ButtonBar>
            <Input
              onChange={onChange}
              className="sm:w-max"
              id="add_key"
              placeholder={placeholder}
              value={value}
            />
            <Button onClick={onClick} variant="outline" icon>
              <X />
            </Button>
          </ButtonBar>
        </div>
        {rowFilled && (
          <div className="flex flex-row py-2">
            <div className="m-2 w-32"></div>
            <div className="flex justify-start">
              <ButtonBar>
                <Button variant="outline" icon>
                  <Plus />
                  {/* onClick={newInput} */}
                </Button>
              </ButtonBar>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

const GWInputButtonList: FC<InputButtonProps> = ({
  value,
  children,
  placeholder,
  onChange,
  rowFilled,
}) => {
  const [inputvalue, setValue] = useState<string>(value ? value : "");

  const onChangeDoSomething = (event: ChangeEvent<HTMLInputElement>) => {
    setValue(event.target.value);
    //  updateList(0, event.target.value);
    rowFilled = true;
  };

  const emptyInput = () => {
    setValue("");
    rowFilled = false;
  };
  const elements: Filter[] = [{ name: "Example" }, { name: "Example2" }];
  /*
  const updateList = (index: number, val: string) => {
    elements[index].name = val;
  };
*/
  return (
    <div className="flex flex-col p-2">
      <p>Test</p>
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
      {elements.map((element) => (
        <div key={element.name} className="flex justify-start">
          <p>{element.name}</p>
        </div>
      ))}
      <div className="flex flex-row py-2">
        <div className="flex justify-center">
          <label
            htmlFor="add_variable"
            className="m-2 w-32 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
          >
            {children}
          </label>
        </div>
        {elements.map((element) => (
          <div key={element.name} className="flex justify-start">
            <ButtonBar>
              <GWInputButton
                onChange={onChangeDoSomething}
                placeholder={element.name}
                value={inputvalue}
                onClick={emptyInput}
              ></GWInputButton>
            </ButtonBar>
          </div>
        ))}
      </div>
      {rowFilled && (
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
      )}
    </div>
  );
};

type Filter = {
  name: string;
};

/*
Before:

          <ButtonBar>
            <Input
              onChange={onChangeDoSomething}
              className="sm:w-max"
              id="add_key"
              placeholder={placeholder}
              value={value2}
            />
            <Button onClick={emptyInput} variant="outline" icon>
              <X />
            </Button>
          </ButtonBar>
*/

export {
  Filepicker,
  GWCheckbox,
  GWForm,
  GWInput2,
  GWSelect,
  GWInputButton,
  GWInputButtonList,
};
