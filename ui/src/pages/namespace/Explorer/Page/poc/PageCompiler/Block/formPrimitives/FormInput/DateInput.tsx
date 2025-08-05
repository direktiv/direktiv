import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { CalendarIcon } from "lucide-react";
import { Datepicker } from "~/design/Datepicker";
import { InputProps } from "./types";
import { useState } from "react";

type DateInputProps = InputProps;

export const DateInput = ({ id, defaultValue }: DateInputProps) => {
  const [date, setDate] = useState<Date | undefined>(defaultValue);
  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="outline" className="w-[300px]">
          <CalendarIcon />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto">
        <Datepicker
          id={id}
          mode="single"
          selected={date}
          // onSelect={setDate}
          initialFocus
        />
      </PopoverContent>
    </Popover>
  );
};
