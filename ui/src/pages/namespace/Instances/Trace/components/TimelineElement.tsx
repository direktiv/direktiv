import { FC } from "react";
import { twMergeClsx } from "~/util/helpers";

type TimelineElementProps = { start: number; end: number; label: string };

const TimelineElement: FC<TimelineElementProps> = ({ start, end, label }) => (
  <div
    className="relative flex flex-row w-full"
    style={{ paddingLeft: `${start}%`, paddingRight: `${end}%` }}
  >
    <div
      className={twMergeClsx(
        "bg-primary-200 dark:bg-primary-700 rounded-sm h-5",
        "overflow-x-visible text-nowrap min-w-px w-full max-w-full"
      )}
      style={start > 50 ? { direction: "rtl" } : {}}
    >
      <div style={{ direction: "ltr", display: "inline-block" }}>{label}</div>
    </div>
  </div>
);

export default TimelineElement;
