import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/design/Command";

import { Search } from "lucide-react";

type Option<T> = {
  value: T;
  label: string;
};

export const SelectFieldMenu = <T extends string>({
  options,
  onSelect,
  heading,
  placeholder,
}: {
  onSelect: (value: T) => void;
  options: Array<Option<T>>;
  heading: string;
  placeholder: string;
}) => (
  <Command>
    <CommandInput icon={Search} placeholder={placeholder} />
    <CommandList>
      <CommandGroup heading={heading}>
        {options.map((option) => (
          <CommandItem
            key={option.value}
            onSelect={() => onSelect(option.value)}
          >
            {option.label}
          </CommandItem>
        ))}
      </CommandGroup>
    </CommandList>
  </Command>
);
