import { FC, PropsWithChildren, useState } from "react";
import { Plus, X } from "lucide-react";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import Input from "~/design/Input";
import { twMergeClsx } from "~/util/helpers";

type FieldsetProps = PropsWithChildren & {
  label: string;
  htmlFor?: string;
  className?: string;
  horizontal?: boolean;
};

export const Fieldset: FC<FieldsetProps> = ({
  label,
  htmlFor,
  children,
  className,
  horizontal,
}) => (
  <fieldset
    className={twMergeClsx(
      "mb-2 flex gap-2",
      className,
      horizontal ? "flex-row-reverse items-center" : "flex-col"
    )}
  >
    <label className="grow text-sm" htmlFor={htmlFor}>
      {label}
    </label>
    {children}
  </fieldset>
);

type ArrayInputProps = {
  placeholder?: string;
  defaultValue: string[];
  onChange: (newValue: string[]) => void;
};

export const ArrayInput: FC<ArrayInputProps> = ({
  defaultValue,
  onChange,
  placeholder,
}) => {
  const [stringArr, setStringArr] = useState(defaultValue);
  const [inputVal, setInputVal] = useState("");

  const addToArray = () => {
    if (inputVal.length > 0) {
      const newStringArr = [...stringArr, inputVal];
      const newStringArrEmptyRemoved = newStringArr.filter(Boolean);
      setInputVal("");
      setStringArr(newStringArrEmptyRemoved);
      onChange(newStringArrEmptyRemoved);
    }
  };

  const changeEntry = (index: number, entry: string) => {
    const newStringArr = stringArr.map((oldValue, oldValueIndex) => {
      if (oldValueIndex === index) {
        return entry;
      }
      return oldValue;
    });
    if (entry) {
      onChange(newStringArr);
    }
    setStringArr(newStringArr);
  };

  const removeEntry = (index: number) => {
    const newStringArr = stringArr.filter(
      (_, oldValueIndex) => oldValueIndex !== index
    );
    const newValueRemovedEmpty = newStringArr.filter(Boolean);
    setStringArr(newValueRemovedEmpty);
    onChange(newValueRemovedEmpty);
  };

  return (
    <div className="grid grid-cols-2 gap-5">
      {stringArr.map((value, valueIndex) => (
        <ButtonBar key={valueIndex}>
          <Input
            placeholder={placeholder}
            value={value}
            onChange={(e) => {
              changeEntry(valueIndex, e.target.value);
            }}
          />
          <Button
            icon
            variant="outline"
            type="button"
            onClick={() => {
              removeEntry(valueIndex);
            }}
          >
            <X />
          </Button>
        </ButtonBar>
      ))}

      <ButtonBar>
        <Input
          placeholder={placeholder}
          value={inputVal}
          onChange={(e) => {
            setInputVal(e.target.value);
          }}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              addToArray();
              e.preventDefault();
            }
          }}
        />
        <Button
          icon
          disabled={!inputVal}
          variant="outline"
          onClick={() => {
            addToArray();
          }}
          type="button"
        >
          <Plus />
        </Button>
      </ButtonBar>
    </div>
  );
};
