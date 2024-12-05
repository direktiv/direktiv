import {
  TimePickerType,
  getTimeByIncrementAndType,
  getTimeFromDate,
  updateDateByTime,
} from "./utils";

import Input from "../Input";
import React from "react";
import { twMergeClsx } from "~/util/helpers";

// this component is mostly copied from https://time.openstatus.dev/

export interface TimePickerInputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {
  picker: TimePickerType;
  date: Date | undefined;
  setDate: (date: Date) => void;
  onRightFocus?: () => void;
  onLeftFocus?: () => void;
}

const TimePickerInput = React.forwardRef<
  HTMLInputElement,
  TimePickerInputProps
>(
  (
    {
      className,
      type = "tel",
      value,
      id,
      name,
      date = new Date(new Date().setHours(0, 0, 0, 0)),
      setDate,
      onChange,
      onKeyDown,
      picker,
      onLeftFocus,
      onRightFocus,
      ...props
    },
    ref
  ) => {
    const [flag, setFlag] = React.useState<boolean>(false);

    /**
     * allow the user to enter the second digit within 2 seconds
     * otherwise start again with entering first digit
     */
    React.useEffect(() => {
      if (flag) {
        const timer = setTimeout(() => {
          setFlag(false);
        }, 2000);

        return () => clearTimeout(timer);
      }
    }, [flag]);

    const calculatedValue = React.useMemo(
      () => getTimeFromDate(date, picker),
      [date, picker]
    );

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Tab") return;
      if (e.key === "ArrowRight") onRightFocus?.();
      if (e.key === "ArrowLeft") onLeftFocus?.();
      if (["ArrowUp", "ArrowDown"].includes(e.key)) {
        const increment = e.key === "ArrowUp" ? 1 : -1;
        const newValue = getTimeByIncrementAndType(
          calculatedValue,
          increment,
          picker
        );
        if (flag) setFlag(false);
        const tempDate = new Date(date);
        setDate(updateDateByTime(tempDate, newValue, picker));
      }
      if (e.key >= "0" && e.key <= "9") {
        const newValue = !flag
          ? "0" + e.key
          : calculatedValue.slice(1, 2) + e.key;
        if (flag) onRightFocus?.();
        setFlag((prev) => !prev);
        const tempDate = new Date(date);
        setDate(updateDateByTime(tempDate, newValue, picker));
      }
    };

    return (
      <Input
        ref={ref}
        id={id || picker}
        name={name || picker}
        className={twMergeClsx(
          "w-[48px] text-center text-base tabular-nums caret-transparent [&::-webkit-inner-spin-button]:appearance-none",
          className
        )}
        value={value || calculatedValue}
        onChange={(e) => {
          e.preventDefault();
          onChange?.(e);
        }}
        type={type}
        inputMode="decimal"
        onKeyDown={(e) => {
          onKeyDown?.(e);
          handleKeyDown(e);
        }}
        {...props}
      />
    );
  }
);

TimePickerInput.displayName = "TimePickerInput";

export default TimePickerInput;
