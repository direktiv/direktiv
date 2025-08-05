import { FormInputType } from "../../../../schema/blocks/form/input";
import Input from "~/design/Input";
import { InputProps } from "./types";

type TextInputProps = InputProps & {
  variant: FormInputType["variant"];
};

export const TextInput = ({ variant, id, defaultValue }: TextInputProps) => (
  <Input type={variant} defaultValue={defaultValue} id={id} />
);
