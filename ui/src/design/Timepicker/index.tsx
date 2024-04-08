import * as React from "react";

import { FC, KeyboardEventHandler } from "react";

import { TimePickerInput } from "./input";

const padTime = (date: number) => date.toString().padStart(2, "0");

export const getTimeString = (date: Date) => {
  const hours = padTime(date.getHours());
  const minutes = padTime(date.getMinutes());
  const seconds = padTime(date.getSeconds());

  const time = hours + ":" + minutes + ":" + seconds;

  return time;
};

type TimePickerProps = {
  date: Date;
  setDate: (date: Date) => void;
  hoursLabel: string;
  minutesLabel: string;
  secondsLabel: string;
  onKeyDown?: KeyboardEventHandler<HTMLInputElement> | undefined;
};

const TimePicker: FC<TimePickerProps> = ({
  date,
  setDate,
  hoursLabel: hours,
  minutesLabel: minutes,
  secondsLabel: seconds,
  onKeyDown,
}) => {
  const hourRef = React.useRef<HTMLInputElement>(null);
  const minuteRef = React.useRef<HTMLInputElement>(null);
  const secondRef = React.useRef<HTMLInputElement>(null);

  return (
    <>
      <div className="flex items-end gap-2 p-2">
        <div className="grid gap-1 text-center">
          <label htmlFor="hours" className="text-xs">
            {hours}
          </label>

          <TimePickerInput
            picker="hours"
            date={date}
            setDate={setDate}
            ref={hourRef}
            onRightFocus={() => minuteRef.current?.focus()}
            onKeyDown={onKeyDown}
          />
        </div>
        <div className="grid gap-1 text-center">
          <label htmlFor="minutes" className="text-xs">
            {minutes}
          </label>
          <TimePickerInput
            picker="minutes"
            date={date}
            setDate={setDate}
            ref={minuteRef}
            onLeftFocus={() => hourRef.current?.focus()}
            onRightFocus={() => secondRef.current?.focus()}
            onKeyDown={onKeyDown}
          />
        </div>
        <div className="grid gap-1 text-center">
          <label htmlFor="seconds" className="text-xs">
            {seconds}
          </label>
          <TimePickerInput
            picker="seconds"
            date={date}
            setDate={setDate}
            ref={secondRef}
            onLeftFocus={() => minuteRef.current?.focus()}
            onKeyDown={onKeyDown}
          />
        </div>
      </div>
    </>
  );
};

export default TimePicker;
