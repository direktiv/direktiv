import { FormInputType } from "../../schema/blocks/form/input";
import { TemplateString } from "../primitives/TemplateString";

type FormInputProps = {
  blockProps: FormInputType;
};

export const FormInput = ({ blockProps }: FormInputProps) => (
  <div>
    <label>
      <TemplateString value={blockProps.label} />
    </label>
    <input type={blockProps.variant} defaultValue={blockProps.defaultValue} />
    <p>
      <TemplateString value={blockProps.description} />
    </p>
  </div>
);
