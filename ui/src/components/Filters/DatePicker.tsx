import { Command, CommandGroup, CommandList } from "~/design/Command";

import { Datepicker } from "~/design/Datepicker";

const DatePicker = ({
  date,
  onChange,
  heading,
}: {
  date?: Date;
  onChange: (value: Date) => void;
  heading: string;
}) => (
  // const setDate = (type: string, value: Date) => {
  // setFilter({
  //   [type]: { type, value },
  // });
  // };

  <Command>
    <CommandList className="max-h-[460px]">
      <CommandGroup heading={heading}>
        <Datepicker
          mode="single"
          selected={date}
          onSelect={(value) => value && onChange(value)}
        />
      </CommandGroup>
    </CommandList>
  </Command>
);
export default DatePicker;
