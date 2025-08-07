import { FormStringInputType } from "../../../../schema/blocks/form/stringInput";
import Input from "~/design/Input";
import { InputProps } from "./types";

type TextInputProps = InputProps & {
  variant: FormStringInputType["variant"];
};

export const TextInput = (props: TextInputProps) => <Input {...props} />;
