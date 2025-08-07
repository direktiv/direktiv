import { FormInputType } from "../../../../schema/blocks/form/input";
import Input from "~/design/Input";
import { InputProps } from "./types";

type TextInputProps = InputProps & {
  variant: FormInputType["variant"];
};

export const TextInput = (props: TextInputProps) => <Input {...props} />;
