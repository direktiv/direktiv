import { Checkbox } from "~/design/Checkbox";
import { FormCheckboxType } from "../../schema/blocks/form/checkbox";
import { TemplateString } from "../primitives/TemplateString";

type FormCheckboxProps = {
  blockProps: FormCheckboxType;
};

export const FormCheckbox = ({ blockProps }: FormCheckboxProps) => (
  <div>
    <label>
      <Checkbox defaultChecked={blockProps.defaultValue} />
      <TemplateString value={blockProps.label} />
    </label>
    <p>
      <TemplateString value={blockProps.description} />
    </p>
  </div>
);
