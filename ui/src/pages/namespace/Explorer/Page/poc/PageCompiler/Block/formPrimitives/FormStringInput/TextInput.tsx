import { FormStringInputType } from "../../../../schema/blocks/form/stringInput";
import Input from "~/design/Input";
import { InputProps } from "./types";

type TextInputProps = InputProps & {
  variant: FormStringInputType["variant"];
};

export const TextInput = (props: TextInputProps) => {
  const { id, defaultValue, variant } = props;
  return <Input id={id} defaultValue={defaultValue} type={variant} />;
};
