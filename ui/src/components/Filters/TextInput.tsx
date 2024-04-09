import { Command, CommandGroup, CommandList } from "~/design/Command";

import { ArrowRight } from "lucide-react";
import Button from "~/design/Button";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import { useState } from "react";

type TextInputProps = {
  value?: string;
  onSubmit: (value: string) => void;
  heading: string;
  placeholder: string;
};

const TextInput = ({
  value,
  onSubmit,
  heading,
  placeholder,
}: TextInputProps) => {
  const [inputValue, setInputValue] = useState<string>(value || "");

  const handleKeyDown = (event: { key: string }) => {
    if (event.key === "Enter") {
      onSubmit(inputValue);
    }
  };

  return (
    <Command>
      <CommandList>
        <CommandGroup heading={heading}>
          <InputWithButton>
            <Input
              autoFocus
              placeholder={placeholder}
              value={inputValue}
              onChange={(event) => setInputValue(event.target.value)}
              onKeyDown={handleKeyDown}
            />
            <Button icon variant="ghost" onClick={() => onSubmit(inputValue)}>
              <ArrowRight />
            </Button>
          </InputWithButton>
        </CommandGroup>
      </CommandList>
    </Command>
  );
};

export default TextInput;
