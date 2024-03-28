import * as React from "react";

import { FC, KeyboardEventHandler } from "react";

import { TimePickerInput } from "./timepicker-input";

export function getTimeString(date: Date) {
  const hours =
    date.getHours() < 10
      ? "0" + String(date.getHours())
      : String(date.getHours());
  const minutes =
    date.getMinutes() < 10
      ? "0" + String(date.getMinutes())
      : String(date.getMinutes());

  const seconds =
    date.getSeconds() < 10
      ? "0" + String(date.getSeconds())
      : String(date.getSeconds());
  const time = hours + ":" + minutes + ":" + seconds;

  return time;
}

type TimePickerProps = {
  date: Date;
  setDate: (date: Date) => void;
  hours: string;
  minutes: string;
  seconds: string;
  onKeyDown?: KeyboardEventHandler<HTMLInputElement> | undefined;
};

const TimePicker: FC<TimePickerProps> = ({
  date,
  setDate,
  hours,
  minutes,
  seconds,
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
