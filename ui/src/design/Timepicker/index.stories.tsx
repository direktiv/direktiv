import { ArrowRight, Plus, X } from "lucide-react";
import { Command, CommandGroup, CommandList } from "~/design/Command";

import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import TimePicker, { showOnlyTimeOfDate } from "./";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Datepicker } from "../Datepicker";

import Input from "~/design/Input";

import { InputWithButton } from "~/design/InputWithButton";
import type { Meta } from "@storybook/react";
import React from "react";
import { format } from "date-fns";

const meta = {
  title: "Components/Timepicker",
  component: TimePicker,
} satisfies Meta<typeof TimePicker>;

export default meta;

export const Default = () => {
  const [date, setDate] = React.useState<Date>(new Date());
  const time = showOnlyTimeOfDate(date);
  return <TimePicker date={date} setDate={setDate} time={time} />;
};

export const TimepickerWithTextinput = () => {
  const [date, setDate] = React.useState<Date>(new Date());
  const time = showOnlyTimeOfDate(date);
  const [name, setName] = React.useState<string>(() => "filename.yaml");

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
                  <TimePicker time={time} date={date} setDate={setDate} />
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

export const ButtonBarWithTimepicker = () => {
  const defaultDate = new Date();

  const [date, setDate] = React.useState<Date>(() => defaultDate);

  const time = showOnlyTimeOfDate(date);
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
                  <TimePicker time={time} date={date} setDate={setDate} />
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
