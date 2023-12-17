import {
  FC,
  Fragment,
  MouseEventHandler,
  PropsWithChildren,
  useState,
} from "react";
import { Plus, X } from "lucide-react";
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
import { Textarea } from "../TextArea";

type FilepickerPropsType = PropsWithChildren & {
  buttonText: string;
  placeholder?: string;
  onChange?: (newValue: string) => void;
  onClick?: MouseEventHandler;
  inputValue?: string;
  displayValue?: string;
};

const GatewayFilepicker: FC<FilepickerPropsType> = ({
  children,
  buttonText,
  placeholder,
  onChange,
  onClick,
  inputValue,
  displayValue,
}) => (
  <div className="flex flex-row py-2">
    <div className="flex justify-center">
      <label
        htmlFor="add_variable"
        className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
      >
        {children}
      </label>
    </div>
    <div className="flex justify-center">
      <ButtonBar>
        <Input
          placeholder={placeholder}
          value={inputValue}
          className="sm:w-max"
          id="add_key"
          onChange={(e) => {
            onChange?.(e.target.value);
          }}
        />

        <Button icon onClick={onClick}>
          {buttonText}
        </Button>
      </ButtonBar>
      <p className="m-2 text-sm">{displayValue}</p>
    </div>
  </div>
);

type CheckboxPropsType = PropsWithChildren & {
  className?: string;
  checked?: boolean;
  onChange?: MouseEventHandler;
};

const GatewayCheckbox: FC<CheckboxPropsType> = ({
  children,
  checked,
  onChange,
}) => (
  <div className="flex flex-row py-2">
    <div className="flex items-center justify-center">
      <label
        htmlFor="GWCheckbox"
        className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
      >
        {children}
      </label>
    </div>
    <div className="flex items-center">
      <Checkbox onClick={onChange} checked={checked} id="GWCheckbox" />
    </div>
  </div>
);

type SelectPropsType = PropsWithChildren & {
  onValueChange?: React.Dispatch<React.SetStateAction<string>>;
  placeholder?: string;
  data: string[];
  value?: string;
};

const GatewaySelect: FC<SelectPropsType> = ({
  children,
  onValueChange,
  placeholder,
  data,
  value,
}) => (
  <div className="flex flex-col py-2 sm:flex-row">
    <label
      htmlFor="select_namespace"
      className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
    >
      {children}
    </label>
    <Select onValueChange={onValueChange}>
      <SelectTrigger value={value} variant="outline">
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          {data.map((element) => (
            <Fragment key={element}>
              <SelectItem value={element}>{element}</SelectItem>
            </Fragment>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  </div>
);

type InputPropsType = PropsWithChildren & {
  className?: string;
  value?: string;
  placeholder?: string;
  onChange?: (newValue: string) => void;
};

const GatewayInput: FC<InputPropsType> = ({
  value,
  children,
  placeholder,
  onChange,
}) => (
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

const GatewayTextarea: FC<InputPropsType> = ({
  value,
  children,
  placeholder,
  onChange,
}) => (
  <div className="flex flex-col py-2 sm:flex-row">
    <label
      htmlFor="add_key"
      className="m-2 w-40 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
    >
      {children}
    </label>
    <Textarea
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

type GatewayArrayProps = PropsWithChildren & {
  placeholder?: string;
  externalArray: string[];
  onChange: (newValue: string[]) => void;
};

const GatewayArray: FC<GatewayArrayProps> = ({
  externalArray,
  onChange,
  children,
  placeholder,
}) => {
  const [internalArray, setInternalArray] = useState(externalArray);
  const [inputVal, setInputVal] = useState("");

  const newValue = (val: string) => {
    if (val.length) {
      setInternalArray((old) => {
        const newValue = [...old, inputVal];
        const newValueRemovedEmpty = newValue.filter(Boolean);
        onChange(newValueRemovedEmpty);
        setInputVal("");
        return newValueRemovedEmpty;
      });
    }
  };

  const changeValue = (valueIndex: number, newValue: string) => {
    setInternalArray((oldArray) => {
      const newArray = oldArray.map((oldValue, index) => {
        if (index === valueIndex) {
          return newValue;
        }
        return oldValue;
      });

      if (newValue) {
        onChange(newArray);
      }
      return newArray;
    });
  };

  const deleteValue = (valueIndex: number) => {
    setInternalArray((old) => {
      const newValue = old.filter((_, i) => i !== valueIndex);
      const newValueRemovedEmpty = newValue.filter(Boolean);
      onChange(newValueRemovedEmpty);
      return newValueRemovedEmpty;
    });
  };

  return (
    <div className="flex flex-row">
      <div className="flex flex-col justify-start">
        <label
          htmlFor="add_variable"
          className="m-2 w-40 py-2 text-sm font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          {children}
        </label>
      </div>
      <div className="flex flex-col">
        {internalArray.map((value, valueIndex) => (
          <div key={valueIndex} className="flex justify-start py-2">
            <ButtonBar>
              <Input
                placeholder={placeholder}
                value={value}
                onChange={(e) => {
                  changeValue(valueIndex, e.target.value);
                }}
                className="sm:w-max"
                id="add_key"
              />
              {}
              <Button
                icon
                variant="outline"
                onClick={() => {
                  deleteValue(valueIndex);
                }}
              >
                <X />
              </Button>
            </ButtonBar>
          </div>
        ))}

        <div className="flex justify-start py-2">
          <ButtonBar>
            <Input
              placeholder={placeholder}
              value={inputVal}
              onChange={(e) => {
                setInputVal(e.target.value);
              }}
            />
            <Button
              icon
              variant="outline"
              onClick={() => {
                newValue(inputVal);
              }}
            >
              <Plus />
            </Button>
          </ButtonBar>
        </div>
      </div>
    </div>
  );
};

export {
  GatewayFilepicker,
  GatewayArray,
  GatewayCheckbox,
  GatewayInput,
  GatewaySelect,
  GatewayTextarea,
};
