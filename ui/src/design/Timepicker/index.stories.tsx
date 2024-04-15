import { ArrowRight, Plus, X } from "lucide-react";
import { Command, CommandGroup, CommandList } from "~/design/Command";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";
import TimePicker, { getTimeString } from "./";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Datepicker } from "../Datepicker";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import type { Meta } from "@storybook/react";
import { format } from "date-fns";
import { useState } from "react";

const meta = {
  title: "Components/Timepicker",
  component: TimePicker,
} satisfies Meta<typeof TimePicker>;

export default meta;

export const Default = () => {
  const [date, setDate] = useState<Date>(new Date());

  return (
    <TimePicker
      hoursLabel="Hours"
      minutesLabel="Minutes"
      secondsLabel="Seconds"
      date={date}
      setDate={setDate}
    />
  );
};

export const TimepickerInButtonBar = () => {
  const [date, setDate] = useState<Date>(new Date());
  const time = getTimeString(date);
  const [name, setName] = useState<string>(() => "filename.yaml");

  return (
    <div className="m-2 flex flex-row flex-wrap gap-2">
      <ButtonBar>
        <Button variant="outline" asChild>
          <label>name</label>
        </Button>
        <Popover>
          <PopoverTrigger asChild>
            <Button variant="outline">
              <span>{name}</span>
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-auto">
            <Command>
              <CommandList>
                <CommandGroup heading="filter by name">
                  <InputWithButton>
                    <Input
                      autoFocus
                      placeholder="filename.yaml"
                      value={name}
                      onChange={(event) => setName(event.target.value)}
                    />
                    <Button icon variant="ghost">
                      <ArrowRight />
                    </Button>
                  </InputWithButton>
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>
        <Button variant="outline" icon>
          <X />
        </Button>
      </ButtonBar>

      <ButtonBar>
        <Button variant="outline" asChild>
          <label>created after</label>
        </Button>
        <Popover>
          <PopoverTrigger asChild>
            <Button variant="outline">
              <span>{time}</span>
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-auto">
            <Command>
              <CommandList>
                <CommandGroup heading="filter by time">
                  <TimePicker
                    hoursLabel="Hours"
                    minutesLabel="Minutes"
                    secondsLabel="Seconds"
                    date={date}
                    setDate={setDate}
                  />
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>
        <Button variant="outline" icon>
          <X />
        </Button>
      </ButtonBar>
    </div>
  );
};

export const ButtonBarWithCombinationOfDatepickerAndTimepicker = () => {
  const defaultDate = new Date();

  const [date, setDate] = useState<Date>(() => defaultDate);

  const time = getTimeString(date);
  return (
    <div className="m-2 flex flex-row flex-wrap gap-2">
      <ButtonBar>
        <Button variant="outline" asChild>
          <label>Created after</label>
        </Button>
        <Popover>
          <PopoverTrigger asChild>
            <Button variant="outline">
              {date ? format(date, "PPP") : <span>Pick a date</span>}
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-auto">
            <Command>
              <CommandList>
                <CommandGroup heading="filter by date">
                  <Datepicker
                    mode="single"
                    selected={date}
                    onDayClick={setDate}
                    initialFocus
                  />
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>
        <Popover>
          <PopoverTrigger asChild>
            <Button variant="outline">
              <span>{time}</span>
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-auto">
            <Command>
              <CommandList>
                <CommandGroup heading="filter by time">
                  <TimePicker
                    hoursLabel="Hours"
                    minutesLabel="Minutes"
                    secondsLabel="Seconds"
                    date={date}
                    setDate={setDate}
                  />
                </CommandGroup>
              </CommandList>
            </Command>
          </PopoverContent>
        </Popover>
        <Button variant="outline" icon>
          <X />
        </Button>
      </ButtonBar>
      <Button variant="outline">
        <Plus />
      </Button>
    </div>
  );
};
