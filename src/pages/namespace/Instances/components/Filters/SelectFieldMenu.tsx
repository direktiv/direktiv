import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/design/Command";

import type { FilterField } from ".";

export const SelectFieldMenu = ({
  onSelect,
}: {
  onSelect: (value: FilterField) => void;
}) => (
  <Command>
    <CommandInput placeholder="Type a command or search..." />
    <CommandList>
      <CommandGroup heading="Select filter">
        <CommandItem tabIndex={1} onSelect={() => onSelect("AS")}>
          by name
        </CommandItem>
        <CommandItem onSelect={() => onSelect("STATUS")}>by state</CommandItem>
        <CommandItem onSelect={() => onSelect("TRIGGER")}>
          by invoker
        </CommandItem>
        <CommandItem onSelect={() => onSelect("AFTER")}>
          created after
        </CommandItem>
        <CommandItem onSelect={() => onSelect("BEFORE")}>
          created before
        </CommandItem>
      </CommandGroup>
    </CommandList>
  </Command>
);
