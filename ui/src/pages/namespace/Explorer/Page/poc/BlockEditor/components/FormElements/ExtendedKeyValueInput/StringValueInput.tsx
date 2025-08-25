import Input from "~/design/Input";

type StringValueInputProps = {
  value: string;
  onChange: (value: string) => void;
  onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void;
  placeholder?: string;
};

export const StringValueInput = ({
  value,
  onChange,
  onKeyDown,
  placeholder = "String value",
}: StringValueInputProps) => (
  <Input
    placeholder={placeholder}
    value={value}
    onKeyDown={onKeyDown}
    onChange={(e) => onChange(e.target.value)}
    className="flex-1"
  />
);
