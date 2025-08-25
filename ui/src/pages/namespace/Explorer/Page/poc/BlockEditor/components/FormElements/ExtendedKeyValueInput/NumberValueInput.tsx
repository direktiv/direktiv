import Input from "~/design/Input";

type NumberValueInputProps = {
  value: number;
  onChange: (value: number) => void;
  onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void;
  placeholder?: string;
};

export const NumberValueInput = ({
  value,
  onChange,
  onKeyDown,
  placeholder = "Number value",
}: NumberValueInputProps) => (
  <Input
    placeholder={placeholder}
    type="number"
    value={value.toString()}
    onKeyDown={onKeyDown}
    onChange={(e) => {
      const numValue = parseFloat(e.target.value) || 0;
      onChange(numValue);
    }}
    className="flex-1"
  />
);
