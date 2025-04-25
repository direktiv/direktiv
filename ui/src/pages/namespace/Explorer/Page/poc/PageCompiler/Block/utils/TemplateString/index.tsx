import { TemplateStringType } from "../../../../schema/primitives/templateString";
import { replaceVariablesInTemplateString } from "./utils";

type TemplateStringProps = {
  value: TemplateStringType;
};

export const TemplateString = ({ value }: TemplateStringProps) => (
  <span>
    {replaceVariablesInTemplateString(
      value,
      (variableName) => `{{${variableName}}}`
    )}
  </span>
);
