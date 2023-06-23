import * as React from "react";

import { ChevronLeft, ChevronRight } from "lucide-react";

import { DayPicker } from "react-day-picker";
import { twMergeClsx } from "~/util/helpers";

export type CalendarProps = React.ComponentProps<typeof DayPicker>;

function Datepicker({
  className,
  classNames,
  showOutsideDays = true,
  ...props
}: CalendarProps) {
  return (
    <DayPicker
      showOutsideDays={showOutsideDays}
      className={twMergeClsx("p-3", className)}
      classNames={{
        months: "flex flex-col sm:flex-row space-y-4 sm:space-x-4 sm:space-y-0",
        month: "space-y-4",
        caption: "flex justify-center pt-1 relative items-center",
        caption_label: "text-sm font-medium",
        nav: "space-x-1 flex items-center",
        nav_button: twMergeClsx(
          "inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2  focus-visible:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none ",
          "border border-gray-4 hover:bg-gray-2",
          "dark:border-gray-dark-4 dark:hover:bg-gray-dark-2",
          "h-7 w-7 bg-transparent p-0 opacity-50 hover:opacity-100"
        ),
        nav_button_previous: "absolute left-1",
        nav_button_next: "absolute right-1",
        table: "w-full border-collapse space-y-1",
        head_row: "flex",
        head_cell:
          "text-gray-11 dark:text-gray-dark-11 rounded-md w-9 font-normal text-[0.8rem]",
        row: "flex w-full mt-2",
        cell: "text-center text-sm p-0 relative [&:has([aria-selected])]:bg-gray-4 dark:[&:has([aria-selected])]:bg-gray-dark-4 first:[&:has([aria-selected])]:rounded-l-md last:[&:has([aria-selected])]:rounded-r-md focus-within:relative focus-within:z-20",
        day: twMergeClsx(
          "inline-flex items-center justify-center rounded-md text-sm transition-colors focus-visible:outline-none focus-visible:ring-2  focus-visible:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none ",
          "bg-transparent data-[state=open]:bg-transparent",
          "hover:bg-gray-3",
          "dark:hover:bg-gray-dark-3",
          "h-9 w-9 p-0 font-medium aria-selected:opacity-100 aria-selected:bg-gray-12 dark:aria-selected:bg-gray-dark-12"
        ),
        day_selected: twMergeClsx(
          "bg-gray-8 hover:bg-gray-12 focus:bg-gray-12 text-gray-1 hover:text-gray-1 focus:text-gray-1",
          "dark:bg-gray-dark-8 dark:hover:bg-gray-dark-12 dark:focus:bg-gray-dark-12 dark:text-gray-dark-1 dark:hover:text-gray-dark-1 dark:focus:text-gray-dark-1"
        ),
        day_today: "",
        day_outside: "text-gray-11 dark:text-gray-dark-11 opacity-50",
        day_disabled: "text-gray-11 dark:text-gray-dark-11 opacity-50",
        day_range_middle:
          "aria-selected:bg-gray-3 dark:aria-selected:bg-gray-dark-3 aria-selected:text-gray-11 dark:text-gray-dark-11 rounded-none",
        day_hidden: "invisible",
        ...classNames,
      }}
      components={{
        IconLeft: () => <ChevronLeft className="h-4 w-4" />,
        IconRight: () => <ChevronRight className="h-4 w-4" />,
      }}
      {...props}
    />
  );
}
Datepicker.displayName = "Datepicker";

export { Datepicker };
