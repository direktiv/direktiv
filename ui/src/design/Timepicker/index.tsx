import * as React from "react";

import { TimePickerInput } from "./timepicker-input";

export function showOnlyTimeOfDate(date: Date) {
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

interface TimePickerProps {
  hours: string;
  minutes: string;
  seconds: string;
  date: Date;
  setDate: (date: Date) => void;
  time: string;
}

function TimePicker({ date, setDate, time }: TimePickerProps) {
  const minuteRef = React.useRef<HTMLInputElement>(null);
  const hourRef = React.useRef<HTMLInputElement>(null);
  const secondRef = React.useRef<HTMLInputElement>(null);
  //const dateInit = new Date(new Date().setHours(0, 0, 0, 0));
  // const showIndicator = !!data?.issues.length;

  const date2 = date ?? new Date(new Date().setHours(0, 0, 0, 0));
  time = showOnlyTimeOfDate(date2);

  return (
    <div className="flex items-end gap-2 p-2">
      <div className="grid gap-1 text-center">
        <label htmlFor="hours" className="text-xs">
          Hours
        </label>
        <TimePickerInput
          picker="hours"
          date={date2}
          setDate={setDate}
          ref={hourRef}
          onRightFocus={() => minuteRef.current?.focus()}
        />
      </div>
      <div className="grid gap-1 text-center">
        <label htmlFor="minutes" className="text-xs">
          Minutes
        </label>
        <TimePickerInput
          picker="minutes"
          date={date2}
          setDate={setDate}
          ref={minuteRef}
          onLeftFocus={() => hourRef.current?.focus()}
          onRightFocus={() => secondRef.current?.focus()}
        />
      </div>
      <div className="grid gap-1 text-center">
        <label htmlFor="seconds" className="text-xs">
          Seconds
        </label>
        <TimePickerInput
          picker="seconds"
          date={date2}
          setDate={setDate}
          ref={secondRef}
          onLeftFocus={() => minuteRef.current?.focus()}
        />
      </div>
    </div>
  );
}

export default TimePicker;
