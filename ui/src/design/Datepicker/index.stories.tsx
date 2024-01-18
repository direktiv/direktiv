import { Popover, PopoverContent, PopoverTrigger } from "../Popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../Select";
import { addDays, format } from "date-fns";
import Button from "../Button";
import { Calendar as CalendarIcon } from "lucide-react";
import { Card } from "../Card";
import type { DateRange } from "react-day-picker";
import { Datepicker } from "./index";
import type { Meta } from "@storybook/react";
import React from "react";
import { twMergeClsx } from "~/util/helpers";

const meta = {
  title: "Components/Datepicker",
  component: Datepicker,
} satisfies Meta<typeof Datepicker>;

export default meta;

export const Default = () => {
  const [date, setDate] = React.useState<Date | undefined>();
  return (
    <Card className="flex w-72 justify-center">
      <Datepicker mode="single" selected={date} onSelect={setDate} />
    </Card>
  );
};

export const DatePicker = () => {
  const [date, setDate] = React.useState<Date>();
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" className="w-[300px]">
          <CalendarIcon />
          {date ? format(date, "PPP") : <span>Pick a date</span>}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto">
        <Datepicker
          mode="single"
          selected={date}
          onSelect={setDate}
          initialFocus
        />
      </PopoverContent>
    </Popover>
  );
};

export function DateRangePicker({
  className,
}: React.HTMLAttributes<HTMLDivElement>) {
  const [date, setDate] = React.useState<DateRange | undefined>({
    from: new Date(2022, 0, 20),
    to: addDays(new Date(2022, 0, 20), 20),
  });

  return (
    <div className={twMergeClsx("grid gap-2", className)}>
      <Popover>
        <PopoverTrigger asChild>
          <Button id="date" variant="outline" className="w-[300px]">
            <CalendarIcon />
            {date?.from &&
              (date.to ? (
                <>
                  {format(date.from, "LLL dd, y")} -{" "}
                  {format(date.to, "LLL dd, y")}
                </>
              ) : (
                format(date.from, "LLL dd, y")
              ))}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <Datepicker
            initialFocus
            mode="range"
            defaultMonth={date?.from}
            selected={date}
            onSelect={setDate}
            numberOfMonths={2}
          />
        </PopoverContent>
      </Popover>
    </div>
  );
}
export function DatepickerWithPresets() {
  const [date, setDate] = React.useState<Date>();
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" className="w-[300px]">
          <CalendarIcon className="mr-2 h-4 w-4" />
          {date ? format(date, "PPP") : <span>Pick a date</span>}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-full flex-col space-y-4 p-2">
        <Select
          onValueChange={(value) =>
            setDate(addDays(new Date(), parseInt(value)))
          }
        >
          <SelectTrigger block>
            <SelectValue placeholder="Select" />
          </SelectTrigger>
          <SelectContent className="w-full">
            <SelectItem value="0">Today</SelectItem>
            <SelectItem value="1">Tomorrow</SelectItem>
            <SelectItem value="3">In 3 days</SelectItem>
            <SelectItem value="7">In a week</SelectItem>
          </SelectContent>
        </Select>
        <Datepicker mode="single" selected={date} onSelect={setDate} />
      </PopoverContent>
    </Popover>
  );
}
