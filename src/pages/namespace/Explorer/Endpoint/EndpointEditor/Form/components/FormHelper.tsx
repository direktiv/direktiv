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
      setStringArr((old) => {
        const newValue = [...old, inputVal];
        const newValueRemovedEmpty = newValue.filter(Boolean);
        onChange(newValueRemovedEmpty);
        setInputVal("");
        return newValueRemovedEmpty;
      });
    }
  };

  const changeEntry = (index: number, entry: string) => {
    setStringArr((oldArray) => {
      const newArray = oldArray.map((oldValue, olcValueIndex) => {
        if (olcValueIndex === index) {
          return entry;
        }
        return oldValue;
      });

      if (entry) {
        onChange(newArray);
      }
      return newArray;
    });
  };

  const removeEntry = (index: number) => {
    setStringArr((old) => {
      const newValue = old.filter((_, i) => i !== index);
      const newValueRemovedEmpty = newValue.filter(Boolean);
      onChange(newValueRemovedEmpty);
      return newValueRemovedEmpty;
    });
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
