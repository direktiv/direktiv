import * as React from "react";

import { ChangeEventHandler, FC, KeyboardEventHandler } from "react";

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
  //setTime: (time: string) => React.SetStateAction<string>;
  setTime: (time: string) => void;
  time: string;
  hours: string;
  minutes: string;
  seconds: string;
  onKeyDown?: (event: KeyboardEvent) => void;
  // KeyboardEvent<HTMLInputElement> // this produces error "Type is not generic"
  // onKeyDown?: KeyboardEventHandler; // did not work because it does not do anything
  onTimeChange: (time: string) => void;
  //onChange: (time: string) => void; // hat nur mit ^ funktioniert: ChangeEventHandler | undefined; // maybe? ChangeEventHandler<T> | undefined;
  //  onChange: (date: Date) => getTimeString(date);
  // onChange: (date: Date) => void;
};

const TimePicker: FC<TimePickerProps> = ({
  date,
  setDate,
  hours,
  minutes,
  seconds,
  onTimeChange,
  onKeyDown,
}) => {
  const minuteRef = React.useRef<HTMLInputElement>(null);
  const hourRef = React.useRef<HTMLInputElement>(null);
  const secondRef = React.useRef<HTMLInputElement>(null);

  const time = getTimeString(date);
  const minutesValue = 0;

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
            onChange={() => {
              console.log("change1 " + time);
              onTimeChange(time);
              onKeyDown;
            }}
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
            value={minutesValue}
            onChange={() => {
              console.log("change2 " + time);
              onTimeChange(time);
              //onKeyDown;
            }}
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
            onChange={() => {
              console.log("change3 " + time);
              onTimeChange(time);
              //onKeyDown;
            }}
          />
        </div>
      </div>
    </>
  );
};

export default TimePicker;
