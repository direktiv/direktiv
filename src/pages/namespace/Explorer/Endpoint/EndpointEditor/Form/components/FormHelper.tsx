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
  externalArray: string[];
  onChange: (newValue: string[]) => void;
};

export const ArrayInput: FC<ArrayInputProps> = ({
  externalArray,
  onChange,
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
    <div className="grid grid-cols-2 gap-5">
      {internalArray.map((value, valueIndex) => (
        <ButtonBar key={valueIndex}>
          <Input
            placeholder={placeholder}
            value={value}
            onChange={(e) => {
              changeValue(valueIndex, e.target.value);
            }}
          />
          {}
          <Button
            icon
            variant="outline"
            type="button"
            onClick={() => {
              deleteValue(valueIndex);
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
              newValue(inputVal);
              e.preventDefault();
            }
          }}
        />
        <Button
          icon
          variant={!inputVal ? "outline" : undefined}
          onClick={() => {
            newValue(inputVal);
          }}
          type="button"
        >
          <Plus />
        </Button>
      </ButtonBar>
    </div>
  );
};
